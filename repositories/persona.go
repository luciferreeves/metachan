package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/enums"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllCharacterStubs() ([]characterStub, error) {
	var stubs []characterStub
	if err := DB.Model(&entities.Character{}).Select("mal_id, enriched_at").Scan(&stubs).Error; err != nil {
		return nil, err
	}
	return stubs, nil
}

func UpdateCharacterDetails(malID int, name, nameKanji, url, imageURL, about string, nicknames []string, favorites int, voiceActors []entities.CharacterVoiceActor, animeAppearances []entities.CharacterAnimeAppearance) error {
	var char entities.Character
	if err := DB.Where("mal_id = ?", malID).First(&char).Error; err != nil {
		return err
	}

	char.Name = name
	char.NameKanji = nameKanji
	char.URL = url
	char.ImageURL = imageURL
	char.About = about
	char.Nicknames = nicknames
	char.Favorites = favorites

	if err := DB.Save(&char).Error; err != nil {
		return err
	}

	DB.Where("character_id = ?", char.ID).Delete(&entities.CharacterVoiceActor{})
	for _, cva := range voiceActors {
		va := cva.VoiceActor
		if va == nil {
			continue
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"url", "image", "name"}),
		}).Create(va)
		if va.ID == 0 {
			DB.Where("mal_id = ?", va.MALID).First(va)
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "character_id"}, {Name: "voice_actor_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"language"}),
		}).Create(&entities.CharacterVoiceActor{
			CharacterID:  char.ID,
			VoiceActorID: va.ID,
			Language:     cva.Language,
		})
	}

	DB.Where("character_id = ?", char.ID).Delete(&entities.CharacterAnimeAppearance{})
	for i := range animeAppearances {
		animeAppearances[i].CharacterID = char.ID
	}
	if len(animeAppearances) > 0 {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "character_id"}, {Name: "anime_mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "url", "image_url", "role"}),
		}).Create(&animeAppearances)
	}

	return nil
}

func GetAnimeCharacters[T idType](maptype enums.MappingType, id T) ([]entities.Character, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return nil, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("anime not found")
	}

	var rows []struct {
		CharacterID uint
		Role        string
	}
	DB.Table("anime_characters").
		Select("character_id, role").
		Where("anime_id = ?", anime.ID).
		Scan(&rows)

	var characters []entities.Character
	for _, row := range rows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}
		DB.Preload("VoiceActor").Where("character_id = ?", char.ID).Find(&char.VoiceActors)
		char.Role = row.Role
		characters = append(characters, char)
	}

	return characters, nil
}

func GetAnimeCharacter[T idType](maptype enums.MappingType, id T, characterMALID int) (entities.Character, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return entities.Character{}, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Character{}, err
		}
		return entities.Character{}, errors.New("anime not found")
	}

	var char entities.Character
	if err := DB.
		Preload("VoiceActors.VoiceActor").
		Preload("AnimeAppearances").
		Where("mal_id = ?", characterMALID).
		First(&char).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Character{}, err
		}
		return entities.Character{}, errors.New("character not found")
	}

	var ac animeCharacterRow
	if err := DB.Table("anime_characters").
		Select("role").
		Where("anime_id = ? AND character_id = ?", anime.ID, char.ID).
		Scan(&ac).Error; err == nil {
		char.Role = ac.Role
	}

	return char, nil
}

func loadAnimeCharacters(anime *entities.Anime) {
	var rows []struct {
		CharacterID uint
		Role        string
	}
	DB.Table("anime_characters").
		Select("character_id, role").
		Where("anime_id = ?", anime.ID).
		Scan(&rows)

	for _, row := range rows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}
		DB.Preload("VoiceActor").
			Where("character_id = ?", char.ID).
			Find(&char.VoiceActors)
		char.Role = row.Role
		anime.Characters = append(anime.Characters, char)
	}
}

func SaveAnimeCharacters(animeID uint, characters []entities.Character) error {
	for i := range characters {
		char := &characters[i]

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"url", "image_url", "name"}),
		}).Create(char)
		if char.ID == 0 {
			DB.Where("mal_id = ?", char.MALID).First(char)
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "anime_id"}, {Name: "character_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role"}),
		}).Create(&entities.AnimeCharacter{
			AnimeID:     animeID,
			CharacterID: char.ID,
			Role:        char.Role,
		})

		for _, cva := range char.VoiceActors {
			va := cva.VoiceActor
			if va == nil {
				continue
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "image", "name"}),
			}).Create(va)
			if va.ID == 0 {
				DB.Where("mal_id = ?", va.MALID).First(va)
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "character_id"}, {Name: "voice_actor_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"language"}),
			}).Create(&entities.CharacterVoiceActor{
				CharacterID:  char.ID,
				VoiceActorID: va.ID,
				Language:     cva.Language,
			})
		}
	}
	return nil
}

func SetCharacterEnriched(malID int) error {
	now := time.Now()
	return DB.Model(&entities.Character{}).Where("mal_id = ?", malID).Update("enriched_at", now).Error
}
