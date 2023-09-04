package gin

import "time"

const SecretKey = "secret"
const bcryptCost = 14
const jwtExpirationHours = 24
const secondsInOneMinute = 60
const minutesInOneHour = 60
const hoursInOneDay = 24

const timeoutDuration = 30 * time.Second
const refreshTokenExpiration = 30 * 24 * time.Hour
const RefreshTokenSecret = "your_refresh_token_secret_key"
