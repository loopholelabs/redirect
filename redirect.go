/*
	Copyright 2023 Loophole Labs

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

package redirect

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	"github.com/valyala/tcplisten"
	"time"
)

var (
	ErrDisabled = errors.New("redirect is disabled")
)

type Options struct {
	LogName       string
	Disabled      bool
	ListenAddress string
}

type Redirect struct {
	server  *fasthttp.Server
	logger  *zerolog.Logger
	options *Options
}

func New(options *Options, logger *zerolog.Logger) (*Redirect, error) {
	l := logger.With().Str(options.LogName, "REDIRECT").Logger()

	if options.Disabled {
		l.Warn().Msg("disabled")
		return nil, ErrDisabled
	}

	return &Redirect{
		server: &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				ctx.URI().SetScheme("https")
				ctx.Redirect(ctx.URI().String(), fasthttp.StatusMovedPermanently)
			},
			ReadTimeout:           time.Second,
			WriteTimeout:          time.Second,
			IdleTimeout:           time.Millisecond * 500,
			NoDefaultServerHeader: true,
			NoDefaultDate:         true,
			NoDefaultContentType:  true,
			CloseOnShutdown:       true,
		},
		logger:  &l,
		options: options,
	}, nil
}

func (s *Redirect) Start() error {
	listenConfig := tcplisten.Config{
		DeferAccept: true,
		FastOpen:    true,
	}

	s.logger.Debug().Msgf("starting redirect on %s", s.options.ListenAddress)
	l, err := listenConfig.NewListener("tcp4", s.options.ListenAddress)
	if err != nil {
		return err
	}
	return s.server.Serve(l)
}

func (s *Redirect) Stop() error {
	return s.server.Shutdown()
}
