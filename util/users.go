package util

import (
	"os/user"
	"strconv"

	"github.com/google/logger"
)

const omogenexecGroup = "omogenexec-users"

func OmogenexecGroupId() int {
	group, err := user.LookupGroup(omogenexecGroup)
	if err != nil {
		logger.Fatalf("could not look up %s group: %v", omogenexecGroup, err)
	}
	id, err := strconv.Atoi(group.Gid)
	if err != nil {
		logger.Fatalf("could not convert gid to int: %v", err)
	}
	return id
}
