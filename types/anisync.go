package types

type AniSyncType string

const (
	AniSyncSPECIAL AniSyncType = "SPECIAL"
	AniSyncTV      AniSyncType = "TV"
	AniSyncOVA     AniSyncType = "OVA"
	AniSyncMOVIE   AniSyncType = "MOVIE"
	AniSyncONA     AniSyncType = "ONA"
	AniSyncUNKNOWN AniSyncType = "UNKNOWN"
)
