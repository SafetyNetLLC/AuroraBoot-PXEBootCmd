package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("build-uki", Label("build-uki", "e2e"), func() {
	var resultDir string
	var keysDir string
	var resultFile string
	var image string
	var err error
	var auroraboot *Auroraboot

	BeforeEach(func() {
		kairosVersion := "v3.3.3"
		resultDir, err = os.MkdirTemp("", "auroraboot-build-uki-test-")
		Expect(err).ToNot(HaveOccurred())
		image = fmt.Sprintf("quay.io/kairos/fedora:40-core-amd64-generic-%s-uki", kairosVersion)
		resultFile = filepath.Join(resultDir, fmt.Sprintf("kairos-fedora-40-core-amd64-generic-%s-uki.iso", kairosVersion))

		currentDir, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		keysDir = filepath.Join(currentDir, "assets", "keys")
		Expect(os.MkdirAll(keysDir, 0755)).ToNot(HaveOccurred())
		auroraboot = NewAuroraboot(resultDir, keysDir)
	})

	AfterEach(func() {
		os.RemoveAll(resultDir)
	})

	Describe("single-efi-cmdline", func() {
		BeforeEach(func() {
			By("building the iso with single-efi-cmdline flags set")
			buildISO(auroraboot, image, keysDir, resultDir, resultFile,
				"--single-efi-cmdline", "My Entry: someoption=somevalue",
				"--single-efi-cmdline", "My Other Entry: someoption2=somevalue2")
		})

		It("creates additional .efi and .conf files", func() {
			content := listEfiFiles(auroraboot, resultFile)
			Expect(content).To(MatchRegexp("my_entry.efi"))
			Expect(content).To(MatchRegexp("my_other_entry.efi"))

			content = listConfFiles(auroraboot, resultFile)
			Expect(content).To(MatchRegexp("my_entry.conf"))
			Expect(content).To(MatchRegexp("my_other_entry.conf"))
		})
	})

	Describe("secure-boot-enroll setting in loader.conf", func() {
		When("secure-boot-enroll is not set", func() {
			BeforeEach(func() {
				By("building the iso with secure-boot-enroll not set")
				buildISO(auroraboot, image, keysDir, resultDir, resultFile)
			})

			It("sets the secure-boot-enroll correctly", func() {
				By("checking if the default value for secure-boot-enroll is set")
				content := readLoaderConf(auroraboot, resultFile)
				Expect(content).To(MatchRegexp("secure-boot-enroll if-safe"))
			})
		})

		When("secure-boot-enroll is set", func() {
			BeforeEach(func() {
				By("building the iso with secure-boot-enroll set to manual")
				buildISO(auroraboot, image, keysDir, resultDir, resultFile, "--secure-boot-enroll", "manual")
			})

			It("sets the secure-boot-enroll correctly", func() {
				By("checking if the user value for secure-boot-enroll is set")
				content := readLoaderConf(auroraboot, resultFile)
				Expect(content).To(MatchRegexp("secure-boot-enroll manual"))
			})
		})
	})
})

func buildISO(auroraboot *Auroraboot, image, keysDir, resultDir, resultFile string, additionalArgs ...string) string {
	By(fmt.Sprintf("building the iso from %s", image))
	args := []string{
		"build-uki",
		"--output-dir", resultDir,
		"--public-keys", keysDir,
		"--tpm-pcr-private-key", filepath.Join(keysDir, "tpm2-pcr-private.pem"),
		"--sb-key", filepath.Join(keysDir, "db.key"),
		"--sb-cert", filepath.Join(keysDir, "db.pem"),
		"--output-type", "iso",
	}
	args = append(args, additionalArgs...)
	args = append(args, image)
	out, err := auroraboot.Run(args...)
	Expect(err).ToNot(HaveOccurred(), out)

	By("building the iso")
	_, err = os.Stat(resultFile)
	Expect(err).ToNot(HaveOccurred(), out)

	return out
}

func readLoaderConf(auroraboot *Auroraboot, isoFile string) string {
	return runCommandInIso(auroraboot, isoFile, "cat /tmp/efi/loader/loader.conf")
}

func listEfiFiles(auroraboot *Auroraboot, isoFile string) string {
	return runCommandInIso(auroraboot, isoFile, "ls /tmp/efi/EFI/kairos")
}

func listConfFiles(auroraboot *Auroraboot, isoFile string) string {
	return runCommandInIso(auroraboot, isoFile, "ls /tmp/efi/loader/entries")
}

func runCommandInIso(auroraboot *Auroraboot, isoFile, command string) string {
	By("running command: " + command)
	out, err := auroraboot.ContainerRun("/bin/bash", "-c",
		fmt.Sprintf(`#!/bin/bash
set -e
mkdir -p /tmp/iso /tmp/efi
mount -v -o loop %[1]s /tmp/iso 2>&1 > /dev/null
sleep 2
mount -v -o loop /tmp/iso/efiboot.img /tmp/efi 2>&1 > /dev/null
%[2]s
umount /tmp/efi 2>&1 > /dev/null
sleep 2
umount /tmp/iso 2>&1 > /dev/null
`, isoFile, command))
	Expect(err).ToNot(HaveOccurred(), out)

	return out
}
