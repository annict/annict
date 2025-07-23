# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/stars", type: :request do
  it "未ログイン時は302でリダイレクトすること" do
    character = FactoryBot.create(:character)

    post "/api/internal/stars", params: {
      starrable_type: "Character",
      starrable_id: character.id
    }

    expect(response.status).to eq(302)
  end

  it "ログイン時、Characterをスターできること" do
    user = FactoryBot.create(:user, :with_email_notification)
    character = FactoryBot.create(:character)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/stars", params: {
        starrable_type: "Character",
        starrable_id: character.id
      }
    }.to change { user.character_favorites.count }.by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})
    expect(user.character_favorites.last.character).to eq(character)
  end

  it "ログイン時、Personをスターできること" do
    user = FactoryBot.create(:user, :with_email_notification)
    person = FactoryBot.create(:person)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/stars", params: {
        starrable_type: "Person",
        starrable_id: person.id
      }
    }.to change { user.person_favorites.count }.by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})
    expect(user.person_favorites.last.person).to eq(person)
  end

  it "ログイン時、Organizationをスターできること" do
    user = FactoryBot.create(:user, :with_email_notification)
    organization = FactoryBot.create(:organization)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/stars", params: {
        starrable_type: "Organization",
        starrable_id: organization.id
      }
    }.to change { user.organization_favorites.count }.by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})
    expect(user.organization_favorites.last.organization).to eq(organization)
  end

  it "既にスター済みの場合でも正常に処理されること" do
    user = FactoryBot.create(:user, :with_email_notification)
    character = FactoryBot.create(:character)
    FactoryBot.create(:character_favorite, user:, character:)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/stars", params: {
        starrable_type: "Character",
        starrable_id: character.id
      }
    }.not_to change { user.character_favorites.count }

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})
  end
end
