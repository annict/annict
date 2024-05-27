# typed: false
# frozen_string_literal: true

describe "Api::V1::People" do
  let(:access_token) { create(:oauth_access_token) }
  let!(:person) { create(:person) }

  describe "GET /v1/people" do
    before do
      get api("/v1/people", access_token: access_token.token)
    end

    context "when added no parameters" do
      it "responses 200" do
        expect(response.status).to eq(200)
      end

      it "gets person info" do
        expected_hash = {
          "id" => person.id,
          "name" => person.name,
          "name_kana" => person.name_kana,
          "name_en" => person.name_en,
          "nickname" => person.nickname,
          "nickname_en" => person.nickname_en,
          "gender_text" => person.gender_text,
          "url" => person.url,
          "url_en" => person.url_en,
          "wikipedia_url" => person.wikipedia_url,
          "wikipedia_url_en" => person.wikipedia_url_en,
          "twitter_username" => person.twitter_username,
          "twitter_username_en" => person.twitter_username_en,
          "birthday" => person.birthday&.strftime("%Y-%m-%d"),
          "blood_type" => person.blood_type,
          "height" => person.height,
          "favorite_people_count" => person.favorite_users_count,
          "casts_count" => person.casts_count,
          "staffs_count" => person.staffs_count,
          "prefecture" => nil
        }
        expect(json["people"][0]).to include(expected_hash)
        expect(json["total_count"]).to eq(1)
        expect(json["next_page"]).to eq(nil)
        expect(json["prev_page"]).to eq(nil)
      end
    end
  end
end
