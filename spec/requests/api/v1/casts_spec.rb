# frozen_string_literal: true

describe "Api::V1::Casts" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:cast) { create(:cast) }

  describe "GET /v1/casts" do
    before do
      get api("/v1/casts", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets cast info" do
        expected_hash = {
          "id" => cast.id,
          "name" => cast.name,
          "name_en" => cast.name_en,
          "sort_number" => cast.sort_number,
          "work" => {
            "id" => cast.anime.id,
            "title" => cast.anime.title,
            "title_en" => "",
            "title_kana" => cast.anime.title_kana,
            "media" => cast.anime.media,
            "media_text" => cast.anime.media_text,
            "released_on" => cast.anime.released_at&.strftime("%Y-%m-%d"),
            "released_on_about" => cast.anime.released_at_about,
            "official_site_url" => cast.anime.official_site_url,
            "wikipedia_url" => cast.anime.wikipedia_url,
            "twitter_username" => cast.anime.twitter_username,
            "twitter_hashtag" => cast.anime.twitter_hashtag,
            "syobocal_tid" => "",
            "mal_anime_id" => cast.anime.mal_anime_id&.to_s,
            "images" => {
              "recommended_url" => cast.anime.recommended_image_url,
              "facebook" => {
                "og_image_url" => cast.anime.facebook_og_image_url
              },
              "twitter" => {
                "mini_avatar_url" => cast.anime.twitter_avatar_url(:mini),
                "normal_avatar_url" => cast.anime.twitter_avatar_url(:normal),
                "bigger_avatar_url" => cast.anime.twitter_avatar_url(:bigger),
                "original_avatar_url" => cast.anime.twitter_avatar_url,
                "image_url" => cast.anime.twitter_image_url
              }
            },
            "episodes_count" => cast.anime.episodes_count,
            "watchers_count" => cast.anime.watchers_count,
            "reviews_count" => cast.anime.work_records_with_body_count,
            "no_episodes" => cast.anime.no_episodes?
          },
          "character" => {
            "id" => cast.character.id,
            "name" => cast.character.name,
            "name_kana" => cast.character.name_kana,
            "name_en" => cast.character.name_en,
            "nickname" => cast.character.nickname,
            "nickname_en" => cast.character.nickname_en,
            "birthday" => cast.character.birthday,
            "birthday_en" => cast.character.birthday_en,
            "age" => cast.character.age,
            "age_en" => cast.character.age_en,
            "blood_type" => cast.character.blood_type,
            "blood_type_en" => cast.character.blood_type_en,
            "height" => cast.character.height,
            "height_en" => cast.character.height_en,
            "weight" => cast.character.weight,
            "weight_en" => cast.character.weight_en,
            "nationality" => cast.character.nationality,
            "nationality_en" => cast.character.nationality_en,
            "occupation" => cast.character.occupation,
            "occupation_en" => cast.character.occupation_en,
            "description" => cast.character.description,
            "description_en" => cast.character.description_en,
            "description_source" => cast.character.description_source,
            "description_source_en" => cast.character.description_source_en,
            "favorite_characters_count" => cast.character.favorite_users_count
          },
          "person" => {
            "id" => cast.person.id,
            "name" => cast.person.name,
            "name_kana" => cast.person.name_kana,
            "name_en" => cast.person.name_en,
            "nickname" => cast.person.nickname,
            "nickname_en" => cast.person.nickname_en,
            "gender_text" => cast.person.gender_text,
            "url" => cast.person.url,
            "url_en" => cast.person.url_en,
            "wikipedia_url" => cast.person.wikipedia_url,
            "wikipedia_url_en" => cast.person.wikipedia_url_en,
            "twitter_username" => cast.person.twitter_username,
            "twitter_username_en" => cast.person.twitter_username_en,
            "birthday" => cast.person.birthday&.strftime("%Y-%m-%d"),
            "blood_type" => cast.person.blood_type,
            "height" => cast.person.height,
            "favorite_people_count" => cast.person.favorite_users_count,
            "casts_count" => cast.person.casts.count,
            "staffs_count" => cast.person.staffs.count
          }
        }
        expect(json["casts"][0]).to include(expected_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
