# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/people/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされ、人物情報は更新されないこと" do
    person = FactoryBot.create(:person)
    old_person = person.attributes
    person_params = {
      name: "徳川家康"
    }

    patch "/db/people/#{person.id}", params: {person: person_params}
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(person.name).to eq(old_person["name"])
  end

  it "編集権限のないユーザーがログインしているとき、アクセスできず、人物情報は更新されないこと" do
    user = FactoryBot.create(:registered_user)
    person = FactoryBot.create(:person)
    old_person = person.attributes
    person_params = {
      name: "徳川家康"
    }

    login_as(user, scope: :user)

    patch "/db/people/#{person.id}", params: {person: person_params}
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(person.name).to eq(old_person["name"])
  end

  it "編集権限のあるユーザーがログインしているとき、人物情報を更新できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person)
    old_person = person.attributes
    person_params = {
      name: "徳川家康"
    }

    login_as(user, scope: :user)

    expect(person.name).to eq(old_person["name"])

    patch "/db/people/#{person.id}", params: {person: person_params}
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(person.name).to eq("徳川家康")
  end

  it "編集権限のあるユーザーがログインしているとき、無効なデータで更新に失敗すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person)
    person_params = {
      name: ""
    }

    login_as(user, scope: :user)

    patch "/db/people/#{person.id}", params: {person: person_params}

    expect(response.status).to eq(422)
    expect(response.body).to include("name=\"person[name]\"")
    expect(response.body).to include("value=\"\"")
  end

  it "編集権限のあるユーザーがログインしているとき、存在しない人物の更新を試みるとエラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person_params = {
      name: "徳川家康"
    }

    login_as(user, scope: :user)

    expect {
      patch "/db/people/99999999", params: {person: person_params}
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "編集権限のあるユーザーがログインしているとき、複数のフィールドを更新できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person)
    person_params = {
      name: "徳川家康",
      name_kana: "とくがわいえやす",
      name_en: "Tokugawa Ieyasu",
      nickname: "家康",
      nickname_en: "Ieyasu",
      gender: "male",
      blood_type: "a",
      birthday: "1543-01-31",
      height: 160,
      url: "https://example.com",
      url_en: "https://example.com/en",
      wikipedia_url: "https://ja.wikipedia.org/wiki/徳川家康",
      wikipedia_url_en: "https://en.wikipedia.org/wiki/Tokugawa_Ieyasu",
      twitter_username: "ieyasu",
      twitter_username_en: "ieyasu_en"
    }

    login_as(user, scope: :user)

    patch "/db/people/#{person.id}", params: {person: person_params}
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(person.name).to eq("徳川家康")
    expect(person.name_kana).to eq("とくがわいえやす")
    expect(person.name_en).to eq("Tokugawa Ieyasu")
    expect(person.nickname).to eq("家康")
    expect(person.nickname_en).to eq("Ieyasu")
    expect(person.gender).to eq("male")
    expect(person.blood_type).to eq("a")
    expect(person.birthday).to eq(Date.parse("1543-01-31"))
    expect(person.height).to eq(160)
    expect(person.url).to eq("https://example.com")
    expect(person.url_en).to eq("https://example.com/en")
    expect(person.wikipedia_url).to eq("https://ja.wikipedia.org/wiki/徳川家康")
    expect(person.wikipedia_url_en).to eq("https://en.wikipedia.org/wiki/Tokugawa_Ieyasu")
    expect(person.twitter_username).to eq("ieyasu")
    expect(person.twitter_username_en).to eq("ieyasu_en")
  end
end
