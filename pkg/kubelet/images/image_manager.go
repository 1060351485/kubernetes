/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package images

import (
	"fmt"
	dockerref "github.com/docker/distribution/reference"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog"
	"os"
	"os/exec"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/events"
	"k8s.io/kubernetes/pkg/util/parsers"
)

// imageManager provides the functionalities for image pulling.
type imageManager struct {
	recorder     record.EventRecorder
	imageService kubecontainer.ImageService
	backOff      *flowcontrol.Backoff
	// It will check the presence of the image, and report the 'image pulling', image pulled' events correspondingly.
	puller imagePuller
}

var _ ImageManager = &imageManager{}

// NewImageManager instantiates a new ImageManager object.
func NewImageManager(recorder record.EventRecorder, imageService kubecontainer.ImageService, imageBackOff *flowcontrol.Backoff, serialized bool, qps float32, burst int) ImageManager {
	imageService = throttleImagePulling(imageService, qps, burst)

	var puller imagePuller
	if serialized {
		puller = newSerialImagePuller(imageService)
	} else {
		puller = newParallelImagePuller(imageService)
	}
	return &imageManager{
		recorder:     recorder,
		imageService: imageService,
		backOff:      imageBackOff,
		puller:       puller,
	}
}

// shouldPullImage returns whether we should pull an image according to
// the presence and pull policy of the image.
func shouldPullImage(container *v1.Container, imagePresent bool) bool {
	if container.ImagePullPolicy == v1.PullNever {
		return false
	}

	if container.ImagePullPolicy == v1.PullAlways ||
		(container.ImagePullPolicy == v1.PullIfNotPresent && (!imagePresent)) {
		return true
	}

	return false
}

// records an event using ref, event msg.  log to glog using prefix, msg, logFn
func (m *imageManager) logIt(ref *v1.ObjectReference, eventtype, event, prefix, msg string, logFn func(args ...interface{})) {
	if ref != nil {
		m.recorder.Event(ref, eventtype, event, msg)
	} else {
		logFn(fmt.Sprint(prefix, " ", msg))
	}
}

func RunCmd(wait bool, name string, args...string) (string, string, error){
	cmd1 := exec.Command(name, args...)
	var err error
	var out1 []byte
	if wait {
		_ = cmd1.Run()
		err = cmd1.Wait()
	}else{
		out1, err = cmd1.CombinedOutput()
	}

	if err != nil {
		msg := fmt.Sprintf("[Jiaheng] run cmd %s, %s failed with error: %s", name, args, err)
		return "", msg, ErrImageInspect
	}
	klog.V(0).Infof("[Jiaheng] run cmd %s, %s success: %s", name, args, out1)
	return string(out1), "", nil
}

// EnsureImageExists pulls the image for the specified pod and container, and returns
// (imageRef, error message, error).
func (m *imageManager) EnsureImageExists(pod *v1.Pod, container *v1.Container, pullSecrets []v1.Secret, podSandboxConfig *runtimeapi.PodSandboxConfig) (string, string, error) {
	klog.V(0).Infof("[Jiaheng] EnsureImageExists called")
	logPrefix := fmt.Sprintf("%s/%s", pod.Name, container.Image)
	ref, err := kubecontainer.GenerateContainerRef(pod, container)
	if err != nil {
		klog.Errorf("Couldn't make a ref to pod %v, container %v: '%v'", pod.Name, container.Name, err)
	}

	// If the image contains no tag or digest, a default tag should be applied.
	image, err := applyDefaultImageTag(container.Image)
	if err != nil {
		msg := fmt.Sprintf("Failed to apply default image tag %q: %v", container.Image, err)
		m.logIt(ref, v1.EventTypeWarning, events.FailedToInspectImage, logPrefix, msg, klog.Warning)
		return "", msg, ErrInvalidImageName
	}
	spec := kubecontainer.ImageSpec{Image: image}
	imageRef, err := m.imageService.GetImageRef(spec)

	if image[:6] == "/ipfs/" {
		downloadPath := "/var/tmp/"
		imageName := "1060351485/largeimage"

		klog.V(0).Infof("[Jiaheng] image hash id is: %s", image[6:])

		// image not exist, call ipfs get
		if _, err := os.Stat(downloadPath + image[6:]); os.IsNotExist(err) {
			cmd1 := []string{"/usr/local/bin/ipfs", "get", image[6:], "-o", downloadPath}
			_, msg1, err1 := RunCmd(true, cmd1[0], cmd1[1:]...)
			if err1 != nil {
				return "", msg1, ErrImageInspect
			}
		}

		// if docker already load the image, skip docker load
		cmd4 := []string{"/bin/sh", "-c", "sudo docker inspect --type=image " + imageName}
		_, _, err4 := RunCmd(true, cmd4[0], cmd4[1:]...)
		if err4 != nil {
			// need to load image
			cmd2 := []string{"/bin/sh", "-c", "sudo docker load -i " + downloadPath + image[6:]}
			_, msg2, err := RunCmd(true, cmd2[0], cmd2[1:]...)
			if err != nil {
				return "", msg2, ErrImageInspect
			}
		}

		// get image id(ref) from docker
		cmd3 := []string{"/bin/sh", "-c", "sudo docker inspect --format=\"{{.Id}}\"" + imageName}
		out3, msg3, err3 := RunCmd(false, cmd3[0], cmd3[1:]...)
		if err3 != nil {
			return "", msg3, ErrImageInspect
		}else{
			return out3, "", nil
		}

		//// image not exist, call ipfs get
		//if _, err := os.Stat(downloadPath + image[6:]); os.IsNotExist(err) {
		//	cmd1 := exec.Command("/usr/local/bin/ipfs", "get", image[6:], "-o", downloadPath)
		//	//out1, err1 := cmd1.CombinedOutput()
		//	_ = cmd1.Run()
		//	err2 := cmd1.Wait()
		//	if err2 != nil {
		//		//msg := fmt.Sprintf("[Jiaheng] ipfs get image failed: %v, %v, with error %s", err1, err2, out1)
		//		msg := fmt.Sprintf("[Jiaheng] ipfs get image failed: %v, with error", err2)
		//		return "", msg, ErrImageInspect
		//	}
		//}
		//klog.V(0).Infof("[Jiaheng] ipfs get image pass")
		//
		//cmd2 := exec.Command("/bin/sh", "-c", "sudo docker load -i " + downloadPath + image[6:])
		////out2, err3 := cmd2.CombinedOutput()
		//_ = cmd2.Run()
		//err4 := cmd2.Wait()
		//if err4 != nil {
		//	//msg := fmt.Sprintf("[Jiaheng] docker load image failed: %v, %v, with err: %s", err3, err4, out2)
		//	msg := fmt.Sprintf("[Jiaheng] docker load image failed: %v, with err", err4)
		//	return "", msg, ErrImageInspect
		//}
		//klog.V(0).Infof("[Jiaheng] docker load image pass")
		//
		//cmd3 := exec.Command("/bin/sh", "-c", "sudo docker inspect --format=\"{{.Id}}\"" + imageName)
		//out3, err5 := cmd3.CombinedOutput()
		//if err5 != nil {
		//	msg := fmt.Sprintf("[Jiaheng] docker inspect image failed: %v, with err: %s", err5, out3)
		//	return "", msg, ErrImageInspect
		//}
		//klog.V(0).Infof("[Jiaheng] docker inspect image pass: %s", out3)
		//imageRef = string(out3)
	}

	if err != nil {
		msg := fmt.Sprintf("Failed to inspect image %q: %v", container.Image, err)
		m.logIt(ref, v1.EventTypeWarning, events.FailedToInspectImage, logPrefix, msg, klog.Warning)
		return "", msg, ErrImageInspect
	}
	klog.V(0).Infof("[Jiaheng] image ref is :%s", imageRef)

	present := imageRef != ""
	if !shouldPullImage(container, present) {
		if present {
			msg := fmt.Sprintf("Container image %q already present on machine", container.Image)
			m.logIt(ref, v1.EventTypeNormal, events.PulledImage, logPrefix, msg, klog.Info)
			return imageRef, "", nil
		}
		msg := fmt.Sprintf("Container image %q is not present with pull policy of Never", container.Image)
		m.logIt(ref, v1.EventTypeWarning, events.ErrImageNeverPullPolicy, logPrefix, msg, klog.Warning)
		return "", msg, ErrImageNeverPull
	}

	backOffKey := fmt.Sprintf("%s_%s", pod.UID, container.Image)
	if m.backOff.IsInBackOffSinceUpdate(backOffKey, m.backOff.Clock.Now()) {
		msg := fmt.Sprintf("Back-off pulling image %q", container.Image)
		m.logIt(ref, v1.EventTypeNormal, events.BackOffPullImage, logPrefix, msg, klog.Info)
		return "", msg, ErrImagePullBackOff
	}
	m.logIt(ref, v1.EventTypeNormal, events.PullingImage, logPrefix, fmt.Sprintf("Pulling image %q", container.Image), klog.Info)
	pullChan := make(chan pullResult)
	m.puller.pullImage(spec, pullSecrets, pullChan, podSandboxConfig)
	imagePullResult := <-pullChan
	if imagePullResult.err != nil {
		m.logIt(ref, v1.EventTypeWarning, events.FailedToPullImage, logPrefix, fmt.Sprintf("Failed to pull image %q: %v", container.Image, imagePullResult.err), klog.Warning)
		m.backOff.Next(backOffKey, m.backOff.Clock.Now())
		if imagePullResult.err == ErrRegistryUnavailable {
			msg := fmt.Sprintf("image pull failed for %s because the registry is unavailable.", container.Image)
			return "", msg, imagePullResult.err
		}

		return "", imagePullResult.err.Error(), ErrImagePull
	}
	m.logIt(ref, v1.EventTypeNormal, events.PulledImage, logPrefix, fmt.Sprintf("Successfully pulled image %q", container.Image), klog.Info)
	m.backOff.GC()
	return imagePullResult.imageRef, "", nil
}

// applyDefaultImageTag parses a docker image string, if it doesn't contain any tag or digest,
// a default tag will be applied.
func applyDefaultImageTag(image string) (string, error) {
	if image[:6] == "/ipfs/"{
		return image, nil
	}
	named, err := dockerref.ParseNormalizedNamed(image)
	if err != nil {
		return "", fmt.Errorf("couldn't parse image reference %q: %v", image, err)
	}
	_, isTagged := named.(dockerref.Tagged)
	_, isDigested := named.(dockerref.Digested)
	if !isTagged && !isDigested {
		// we just concatenate the image name with the default tag here instead
		// of using dockerref.WithTag(named, ...) because that would cause the
		// image to be fully qualified as docker.io/$name if it's a short name
		// (e.g. just busybox). We don't want that to happen to keep the CRI
		// agnostic wrt image names and default hostnames.

		image = image + ":" + parsers.DefaultImageTag
	}
	return image, nil
}
