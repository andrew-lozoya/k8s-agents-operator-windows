/*
Copyright 2024.

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
package apm

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/andrew-lozoya/k8s-agents-operator-windows/src/api/v1alpha2"
)

const (
	envDotnetFrameworkWindowsClrEnableProfiling     = "COR_ENABLE_PROFILING"
	envDotnetFrameworkWindowsClrProfiler            = "COR_PROFILER"
	envDotnetFrameworkWindowsClrProfilerPath        = "COR_PROFILER_PATH"
	envDotnetFrameworkWindowsNewrelicHome           = "NEWRELIC_HOME"
	dotnetFrameworkWindowsClrEnableProfilingEnabled = "1"
	dotnetFrameworkWindowsClrProfilerID             = "{71DA0A04-7777-4EC6-9643-7D28B46A8A41}"
	dotnetFrameworkWindowsClrProfilerPath           = "\\newrelic-instrumentation\\netframework\\NewRelic.Profiler.dll"
	dotnetFrameworkWindowsNewrelicHomePath          = "\\newrelic-instrumentation\\netframework"
	dotnetFrameworkWindowsInitContainerName         = initContainerName + "-dotnet-framework-windows"
)

var _ Injector = (*DotnetFrameworkWindowsInjector)(nil)

func init() {
	DefaultInjectorRegistry.MustRegister(&DotnetFrameworkWindowsInjector{})
}

type DotnetFrameworkWindowsInjector struct {
	baseInjector
}

func (i *DotnetFrameworkWindowsInjector) Language() string {
	return "dotnet-framework-windows"
}

func (i *DotnetFrameworkWindowsInjector) acceptable(inst v1alpha2.Instrumentation, pod corev1.Pod) bool {
	if inst.Spec.Agent.Language != i.Language() {
		return false
	}
	if len(pod.Spec.Containers) == 0 {
		return false
	}
	return true
}

func (i DotnetFrameworkWindowsInjector) Inject(ctx context.Context, inst v1alpha2.Instrumentation, ns corev1.Namespace, pod corev1.Pod) (corev1.Pod, error) {
	if !i.acceptable(inst, pod) {
		return pod, nil
	}
	if err := i.validate(inst); err != nil {
		return pod, err
	}

	firstContainer := 0
	// caller checks if there is at least one container.
	container := &pod.Spec.Containers[firstContainer]

	// check if NEWRELIC_HOME env var is already set in the container
	// if it is already set, then we assume that .NET newrelic-instrumentation is already configured for this container
	if getIndexOfEnv(container.Env, envDotnetFrameworkWindowsNewrelicHome) > -1 {
		return pod, errors.New("NEWRELIC_HOME environment variable is already set in the container")
	}

	// check if NEWRELIC_HOME env var is already set in the .NET instrumentation spec
	// if it is already set, then we assume that .NET newrelic-instrumentation is already configured for this container
	if getIndexOfEnv(inst.Spec.Agent.Env, envDotnetFrameworkWindowsNewrelicHome) > -1 {
		return pod, errors.New("NEWRELIC_HOME environment variable is already set in the .NET instrumentation spec")
	}

	// inject .NET instrumentation spec env vars.
	for _, env := range inst.Spec.Agent.Env {
		idx := getIndexOfEnv(container.Env, env.Name)
		if idx == -1 {
			container.Env = append(container.Env, env)
		}
	}

	setEnvVar(container, envDotnetFrameworkWindowsClrEnableProfiling, dotnetFrameworkWindowsClrEnableProfilingEnabled, false)
	setEnvVar(container, envDotnetFrameworkWindowsClrProfiler, dotnetFrameworkWindowsClrProfilerID, false)
	setEnvVar(container, envDotnetFrameworkWindowsClrProfilerPath, dotnetFrameworkWindowsClrProfilerPath, false)
	setEnvVar(container, envDotnetFrameworkWindowsNewrelicHome, dotnetFrameworkWindowsNewrelicHomePath, false)

	if isContainerVolumeMissing(container, volumeName) {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: "\\newrelic-instrumentation",
		})
	}

	// We just inject Volumes and init containers for the first processed container.
	if isInitContainerMissing(pod, dotnetFrameworkWindowsInitContainerName) {
		if isPodVolumeMissing(pod, volumeName) {
			pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				}})
		}

		pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
			Name:    dotnetFrameworkWindowsInitContainerName,
			Image:   inst.Spec.Agent.Image,
			Command: []string{"powershell", "-Command", "Copy-Item -Path \\instrumentation\\* -Destination \\newrelic-instrumentation -Recurse -Force"},
			VolumeMounts: []corev1.VolumeMount{{
				Name:      volumeName,
				MountPath: "\\newrelic-instrumentation",
			}},
		})
	}

	pod = i.injectNewrelicConfig(ctx, inst.Spec.Resource, ns, pod, firstContainer, inst.Spec.LicenseKeySecret)

	return pod, nil
}
