#!/bin/bash -e

curl -s \
	-H "Authorization: Bearer ${ACTIONS_ID_TOKEN_REQUEST_TOKEN}" \
	"${ACTIONS_ID_TOKEN_REQUEST_URL}?audience=oblique" | \
	jq '.value' | \
	awk -F'.' '{print $2}' | \
	tr '_-' '/+' | \
	base64 -d | \
	jq .


