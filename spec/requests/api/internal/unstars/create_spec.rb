# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/unstars", type: :request do
  it "未認証のとき、302ステータスを返すこと" do
    character = create(:character)
    data = {starrable_type: "Character", starrable_id: character.id}

    post api("/api/internal/unstars", data)

    expect(response.status).to eq(302)
  end

  it "キャラクターのスターを外したとき、201ステータスを返すこと" do
    user = create(:registered_user)
    character_favorite = create(:character_favorite, user:)
    data = {starrable_type: "Character", starrable_id: character_favorite.character_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "キャラクターのスターを外したとき、レコードが削除されること" do
    user = create(:registered_user)
    character_favorite = create(:character_favorite, user:)
    data = {starrable_type: "Character", starrable_id: character_favorite.character_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_characters.count }.from(1).to(0)
  end

  it "組織のスターを外したとき、201ステータスを返すこと" do
    user = create(:registered_user)
    organization_favorite = create(:organization_favorite, user:)
    data = {starrable_type: "Organization", starrable_id: organization_favorite.organization_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "組織のスターを外したとき、レコードが削除されること" do
    user = create(:registered_user)
    organization_favorite = create(:organization_favorite, user:)
    data = {starrable_type: "Organization", starrable_id: organization_favorite.organization_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_organizations.count }.from(1).to(0)
  end

  it "人物のスターを外したとき、201ステータスを返すこと" do
    user = create(:registered_user)
    person_favorite = create(:person_favorite, user:)
    data = {starrable_type: "Person", starrable_id: person_favorite.person_id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end

  it "人物のスターを外したとき、レコードが削除されること" do
    user = create(:registered_user)
    person_favorite = create(:person_favorite, user:)
    data = {starrable_type: "Person", starrable_id: person_favorite.person_id}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }
      .to change { user.favorite_people.count }.from(1).to(0)
  end

  it "存在しないstarrable_typeを指定したとき、NameErrorが発生すること" do
    user = create(:registered_user)
    data = {starrable_type: "InvalidType", starrable_id: 1}

    login_as(user, scope: :user)
    expect { post api("/api/internal/unstars", data) }.to raise_error(NameError)
  end

  it "存在しないstarrable_idを指定したとき、404エラーが返されること" do
    user = create(:registered_user)
    data = {starrable_type: "Character", starrable_id: 99999}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response).to have_http_status(:not_found)
  end

  it "スターしていないリソースのスターを外そうとしたとき、201ステータスを返すこと" do
    user = create(:registered_user)
    character = create(:character)
    data = {starrable_type: "Character", starrable_id: character.id}

    login_as(user, scope: :user)
    post api("/api/internal/unstars", data)

    expect(response.status).to eq(201)
  end
end
