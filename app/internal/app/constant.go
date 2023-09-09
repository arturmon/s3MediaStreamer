package app

import "time"

const (
	Quarter             = 3 * 30 * 24 * time.Hour // Quarter (approximately)
	Month               = 30 * 24 * time.Hour     // Month (approximately)
	HalfYear            = 6 * 30 * 24 * time.Hour // HalfYear (approximately)
	Day                 = 24 * time.Hour          // Day
	SixHoursDuration    = 6 * time.Hour           // 6 Hours
	TwelveHoursDuration = 12 * time.Hour          // 12 Hours
	TwoDays             = 2 * 24 * time.Hour      // 2 Days
	ThreeDays           = 3 * 24 * time.Hour      // 3 Days
	OneWeek             = 7 * 24 * time.Hour      // 1 Week
	TwoWeeks            = 14 * 24 * time.Hour     // 2 Weeks
	OneMonth            = 30 * 24 * time.Hour     // 1 Month
	ThreeMonths         = 3 * 30 * 24 * time.Hour // 3 Months
	SixMonths           = 6 * 30 * 24 * time.Hour // 6 Months
	OneYear             = 365 * 24 * time.Hour    // 1 OneYear (approximately)

	minSubmatchesCount = 4
)
