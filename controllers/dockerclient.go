/*
 * Copyright (c) 2020.
 *
 * Metaprov.com
 */

package controllers

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

//////////////////////////////////////////////////////////////////////////////
/////  DockerClient
//////////////////////////////////////////////////////////////////////////////

type DockerClient interface {
	// Build a docker image
	Build(imagename string, fileReader *bytes.Reader) error
	// Push a docker image.
	Push(regaddress string, imagename string, uname string, password string) error
	// Pull a docker image
	Pull(imagename string) error
}

///////////////////////////////////////////////////////////////
///// Real Docker client
//////////////////////////////////////////////////////////////

// The real impl
type RealDockerClient struct {
}

func (r *RealDockerClient) Build(imagename string, fileReader *bytes.Reader) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("Error getting docker client: %s.", err)
	}

	cli.NegotiateAPIVersion(context.Background())

	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		Tags:           []string{imagename},
		Dockerfile:     "Dockerfile",
		PullParent:     false}
	buildResponse, err := cli.ImageBuild(context.Background(), fileReader, options)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return errors.Wrapf(err, " :unable to read image build response")
	}
	return nil
}

func (r *RealDockerClient) Pull(imagename string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("Error getting docker client: %s.", err)
	}

	cli.NegotiateAPIVersion(context.Background())

	options := types.ImagePullOptions{}
	_, err = cli.ImagePull(context.Background(), imagename, options)
	if err != nil {
		return err
	}
	return nil
}

// Push the image to an public docker registry.
// We assume that the image exist.
func (im *RealDockerClient) Push(regaddress string, imagename string, uname string, password string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return errors.Wrapf(err, "error creating client: %v", err)
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	auth := types.AuthConfig{
		Username:      uname,
		Password:      password,
		ServerAddress: regaddress,
	}
	authBytes, err := json.Marshal(auth)
	if err != nil {
		return errors.Wrapf(err, "json.Marshal")
	}
	authBase64 := b64.URLEncoding.EncodeToString(authBytes)

	pushOptions := types.ImagePushOptions{}
	pushOptions.RegistryAuth = authBase64
	pushOptions.All = true

	closer, err := cli.ImagePush(ctx, imagename, pushOptions)
	if err != nil {
		return errors.Wrapf(err, "Image Push %v failed", imagename)
	}

	_, err = io.Copy(os.Stdout, closer)
	if err != nil {
		return errors.Wrapf(err, "io.Copy failed")
	}

	closer.Close()
	return nil
}

//////////////////////////////////////////////////////////////////
///// Tested docker client
/////////////////////////////////////////////////////////////////

type TestDockerClient struct {
	gotImageName string
	pushed       bool
	pulled       bool
}

func (t *TestDockerClient) Build(imagename string, fileReader *bytes.Reader) error {
	t.gotImageName = imagename
	return nil
}

func (t *TestDockerClient) Push(regaddress string, imagename string, uname string, password string) error {
	t.pushed = true
	return nil
}

func (t *TestDockerClient) Pull(imagename string) error {
	t.pulled = true
	return nil
}
