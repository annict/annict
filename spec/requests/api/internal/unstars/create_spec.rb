# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/unstars", type: :request do
  it "キャラクターのスターを外した時、201ステータスを返すこと" do
    user = create(:registered_user)
    character_favorite = create(:character_favorite, user:)
    data = {starrable_type: "Character", starrable_id: character_favorite.character_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "キャラクターのスターを外した時、レコードが削除されること" do
    user = create(:registered_user)
    character_favorite = create(:character_favorite, user:)
    data = {starrable_type: "Character", starrable_id: character_favorite.character_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_characters.count }.from(1).to(0)
  end

  it "組織のスターを外した時、201ステータスを返すこと" do
    user = create(:registered_user)
    organization_favorite = create(:organization_favorite, user:)
    data = {starrable_type: "Organization", starrable_id: organization_favorite.organization_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "組織のスターを外した時、レコードが削除されること" do
    user = create(:registered_user)
    organization_favorite = create(:organization_favorite, user:)
    data = {starrable_type: "Organization", starrable_id: organization_favorite.organization_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_organizations.count }.from(1).to(0)
  end

  it "人物のスターを外した時、201ステータスを返すこと" do
    user = create(:registered_user)
    person_favorite = create(:person_favorite, user:)
    data = {starrable_type: "Person", starrable_id: person_favorite.person_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "人物のスターを外した時、レコードが削除されること" do
    user = create(:registered_user)
    person_favorite = create(:person_favorite, user:)
    data = {starrable_type: "Person", starrable_id: person_favorite.person_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_people.count }.from(1).to(0)
  end
end
