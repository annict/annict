# typed: false
# frozen_string_literal: true

RSpec.describe "GET /v1/staffs", type: :request do
  it "パラメーターを指定しない場合、200ステータスを返すこと" do
    access_token = create(:oauth_access_token)
    create(:staff)

    get api("/v1/staffs", access_token: access_token.token)

    expect(response.status).to eq(200)
  end

  it "パラメーターを指定しない場合、スタッフ情報を取得できること" do
    access_token = create(:oauth_access_token)
    staff = create(:staff)

    get api("/v1/staffs", access_token: access_token.token)

    expected_hash = {
      "id" => staff.id,
      "name" => staff.name,
      "name_en" => staff.name_en,
      "role_text" => staff.role_text,
      "role_other" => staff.role_other,
      "role_other_en" => staff.role_other_en,
      "sort_number" => staff.sort_number,
      "work" => {
        "id" => staff.work.id,
        "title" => staff.work.title,
        "title_en" => "",
        "title_kana" => staff.work.title_kana,
        "media" => staff.work.media,
        "media_text" => staff.work.media_text,
        "released_on" => staff.work.released_at&.strftime("%Y-%m-%d"),
        "released_on_about" => staff.work.released_at_about,
        "official_site_url" => staff.work.official_site_url,
        "wikipedia_url" => staff.work.wikipedia_url,
        "twitter_username" => staff.work.twitter_username,
        "twitter_hashtag" => staff.work.twitter_hashtag,
        "syobocal_tid" => "",
        "mal_anime_id" => staff.work.mal_anime_id&.to_s,
        "images" => {
          "recommended_url" => staff.work.recommended_image_url,
          "facebook" => {
            "og_image_url" => staff.work.facebook_og_image_url
          },
          "twitter" => {
            "mini_avatar_url" => staff.work.twitter_avatar_url(:mini),
            "normal_avatar_url" => staff.work.twitter_avatar_url(:normal),
            "bigger_avatar_url" => staff.work.twitter_avatar_url(:bigger),
            "original_avatar_url" => staff.work.twitter_avatar_url,
            "image_url" => staff.work.twitter_image_url
          }
        },
        "episodes_count" => staff.work.episodes_count,
        "watchers_count" => staff.work.watchers_count,
        "reviews_count" => staff.work.work_records_with_body_count,
        "no_episodes" => staff.work.no_episodes?
      },
      "person" => {
        "id" => staff.resource.id,
        "name" => staff.resource.name,
        "name_kana" => staff.resource.name_kana,
        "name_en" => staff.resource.name_en,
        "nickname" => staff.resource.nickname,
        "nickname_en" => staff.resource.nickname_en,
        "gender_text" => staff.resource.gender_text,
        "url" => staff.resource.url,
        "url_en" => staff.resource.url_en,
        "wikipedia_url" => staff.resource.wikipedia_url,
        "wikipedia_url_en" => staff.resource.wikipedia_url_en,
        "twitter_username" => staff.resource.twitter_username,
        "twitter_username_en" => staff.resource.twitter_username_en,
        "birthday" => staff.resource.birthday&.strftime("%Y-%m-%d"),
        "blood_type" => staff.resource.blood_type,
        "height" => staff.resource.height,
        "favorite_people_count" => staff.resource.favorite_users_count,
        "casts_count" => staff.resource.casts.count,
        "staffs_count" => staff.resource.staffs.count
      }
    }
    expect(json["staffs"][0]).to include(expected_hash)
    expect(json["total_count"]).to eq(1)
    expect(json["next_page"]).to eq(nil)
    expect(json["prev_page"]).to eq(nil)
  end

  it "アクセストークンが無効な場合、401ステータスを返すこと" do
    get api("/v1/staffs", access_token: "invalid_token")

    expect(response.status).to eq(401)
  end

  it "スタッフが存在しない場合、空の配列を返すこと" do
    access_token = create(:oauth_access_token)

    get api("/v1/staffs", access_token: access_token.token)

    expect(response.status).to eq(200)
    expect(json["staffs"]).to eq([])
    expect(json["total_count"]).to eq(0)
  end
end
