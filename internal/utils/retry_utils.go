package utils

import "time"

const RetryMainSleep = 2 * time.Second
const RetrySuccessSleep = 2 * time.Second
const RetryErrorSleep = 5 * time.Second
const RetryExpiredContextSleep = 60 * time.Second
