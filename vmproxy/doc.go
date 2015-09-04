// Copyright 2015 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the APACHE 2.0 License.

/*
Package vmproxy provides tools to proxy App Engine requests to on-demand,
Compute Engine instances.

This package provides simple tool that enables one to install an http.Handler
that will forward requests from App Engine to a Compute Engine VM.

Stateless Batch Jobs

A first use-case for vmproxy is to handle background, sporadic,
stateless batch jobs. Some examples include, daily report generation,
statistics, image processing, vídeo transcoding and others.

*/
package vmproxy