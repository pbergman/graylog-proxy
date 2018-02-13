package net

import (
    "regexp"
)

type GraylogHost struct {
    network string
    host string
    secure bool

}

func (g GraylogHost) String() string {
    return g.network + "://" +g.host
}

func (g GraylogHost) GetNetwork() string {
    return g.network
}

func (g GraylogHost) GetHost() string {
    return g.host
}

func (g GraylogHost) IsSecure() bool {
    return g.secure
}

func NewGraylogHost(host string) *GraylogHost {
    pattern := regexp.MustCompile(`(?:(tcp[4|6]?(?:\+tls)?|http[s]?)://)([^$]+)`)
    if pattern.MatchString(host) {
        info := pattern.FindStringSubmatch(host)
        var network string
        var secure bool
        if info[1][:3] == "tcp" {
            if 3 == len(info[1]) {
                network = info[1]
            } else if '+' == info[1][3] {
                network = info[1][:3]
                secure = true
            } else {
                network = info[1][:4]
            }
        } else {
            network = info[1]
            if "https" == network {
                secure = true
            }
        }
        return &GraylogHost{network, info[2], secure}
    }
    return nil
}