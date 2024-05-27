# typed: false
# frozen_string_literal: true

describe "Api::V1::Characters" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:character) { create(:character) }

  describe "GET /v1/characters" do
    before do
      get api("/v1/characters", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets character info" do
        expected_hash = {
          "id" => character.id,
          "name" => character.name,
          "name_kana" => character.name_kana,
          "name_en" => character.name_en,
          "nickname" => character.nickname,
          "nickname_en" => character.nickname_en,
          "birthday" => character.birthday,
          "birthday_en" => character.birthday_en,
          "age" => character.age,
          "age_en" => character.age_en,
          "blood_type" => character.blood_type,
          "blood_type_en" => character.blood_type_en,
          "height" => character.height,
          "height_en" => character.height_en,
          "weight" => character.weight,
          "weight_en" => character.weight_en,
          "nationality" => character.nationality,
          "nationality_en" => character.nationality_en,
          "occupation" => character.occupation,
          "occupation_en" => character.occupation_en,
          "description" => character.description,
          "description_en" => character.description_en,
          "description_source" => character.description_source,
          "description_source_en" => character.description_source_en,
          "favorite_characters_count" => character.favorite_users_count,
          "series" => {
            "id" => character.series.id,
            "name" => character.series.name,
            "name_en" => character.series.name_en,
            "name_ro" => character.series.name_ro
          }
        }
        expect(json["characters"][0]).to include(expected_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
