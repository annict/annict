# typed: false
# frozen_string_literal: true

RSpec.describe "GET /api/internal/stars", type: :request do
  it "未ログイン時は空の配列を返すこと" do
    get "/api/internal/stars"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end

  it "ログイン時は自分のスターのリストを返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)
    other_user = FactoryBot.create(:user, :with_email_notification)

    # Characterのスターを作成
    character = FactoryBot.create(:character)
    FactoryBot.create(:character_favorite, user:, character:)

    # Personのスターを作成
    person = FactoryBot.create(:person)
    FactoryBot.create(:person_favorite, user:, person:)

    # Organizationのスターを作成
    organization = FactoryBot.create(:organization)
    FactoryBot.create(:organization_favorite, user:, organization:)

    # 他のユーザーのスターを作成（取得されないはず）
    other_character = FactoryBot.create(:character)
    FactoryBot.create(:character_favorite, user: other_user, character: other_character)

    login_as(user, scope: :user)
    get "/api/internal/stars"

    expect(response.status).to eq(200)

    stars = JSON.parse(response.body)
    expect(stars.size).to eq(3)

    # レスポンスの内容を検証
    expected_stars = [
      {
        "starrable_type" => "Character",
        "starrable_id" => character.id
      },
      {
        "starrable_type" => "Person",
        "starrable_id" => person.id
      },
      {
        "starrable_type" => "Organization",
        "starrable_id" => organization.id
      }
    ]

    expect(stars).to match_array(expected_stars)
  end

  it "Characterのスターのみの場合も正しく返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    # Characterのスターのみ作成
    character1 = FactoryBot.create(:character)
    character2 = FactoryBot.create(:character)
    FactoryBot.create(:character_favorite, user:, character: character1)
    FactoryBot.create(:character_favorite, user:, character: character2)

    login_as(user, scope: :user)
    get "/api/internal/stars"

    expect(response.status).to eq(200)

    stars = JSON.parse(response.body)
    expect(stars.size).to eq(2)
    expect(stars).to match_array([
      {"starrable_type" => "Character", "starrable_id" => character1.id},
      {"starrable_type" => "Character", "starrable_id" => character2.id}
    ])
  end

  it "Personのスターのみの場合も正しく返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    # Personのスターのみ作成
    person1 = FactoryBot.create(:person)
    person2 = FactoryBot.create(:person)
    FactoryBot.create(:person_favorite, user:, person: person1)
    FactoryBot.create(:person_favorite, user:, person: person2)

    login_as(user, scope: :user)
    get "/api/internal/stars"

    expect(response.status).to eq(200)

    stars = JSON.parse(response.body)
    expect(stars.size).to eq(2)
    expect(stars).to match_array([
      {"starrable_type" => "Person", "starrable_id" => person1.id},
      {"starrable_type" => "Person", "starrable_id" => person2.id}
    ])
  end

  it "Organizationのスターのみの場合も正しく返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    # Organizationのスターのみ作成
    organization1 = FactoryBot.create(:organization)
    organization2 = FactoryBot.create(:organization)
    FactoryBot.create(:organization_favorite, user:, organization: organization1)
    FactoryBot.create(:organization_favorite, user:, organization: organization2)

    login_as(user, scope: :user)
    get "/api/internal/stars"

    expect(response.status).to eq(200)

    stars = JSON.parse(response.body)
    expect(stars.size).to eq(2)
    expect(stars).to match_array([
      {"starrable_type" => "Organization", "starrable_id" => organization1.id},
      {"starrable_type" => "Organization", "starrable_id" => organization2.id}
    ])
  end

  it "スターがない場合は空の配列を返すこと" do
    user = FactoryBot.create(:user, :with_email_notification)

    login_as(user, scope: :user)
    get "/api/internal/stars"

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq([])
  end
end
