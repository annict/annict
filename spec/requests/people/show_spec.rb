# typed: false
# frozen_string_literal: true

RSpec.describe "GET /people/:person_id", type: :request do
  it "人物が存在する場合、200を返し、人物名が表示されること" do
    person = FactoryBot.create(:person)

    get "/people/#{person.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "削除済みの人物の場合、404エラーになること" do
    person = FactoryBot.create(:person, deleted_at: Time.current)

    expect {
      get "/people/#{person.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない人物IDの場合、404エラーになること" do
    get "/people/nonexistent"

    expect(response.status).to eq(404)
  end

  it "声優の場合、出演作品情報が表示されること" do
    person = FactoryBot.create(:person)
    work1 = FactoryBot.create(:work, season_year: 2023, title: "作品1")
    work2 = FactoryBot.create(:work, season_year: 2022, title: "作品2")
    character1 = FactoryBot.create(:character, name: "キャラクター1")
    character2 = FactoryBot.create(:character, name: "キャラクター2")
    FactoryBot.create(:cast, person:, work: work1, character: character1)
    FactoryBot.create(:cast, person:, work: work2, character: character2)

    get "/people/#{person.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("2023")
    expect(response.body).to include("2022")
    expect(response.body).to include("作品1")
    expect(response.body).to include("作品2")
  end

  it "スタッフの場合、参加作品情報が表示されること" do
    person = FactoryBot.create(:person)
    work1 = FactoryBot.create(:work, season_year: 2023, title: "作品1")
    work2 = FactoryBot.create(:work, season_year: 2022, title: "作品2")
    FactoryBot.create(:staff, resource: person, work: work1, name: "監督")
    FactoryBot.create(:staff, resource: person, work: work2, name: "原作")

    get "/people/#{person.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("2023")
    expect(response.body).to include("2022")
    expect(response.body).to include("作品1")
    expect(response.body).to include("作品2")
  end

  it "削除済み作品は除外されること" do
    person = FactoryBot.create(:person)
    work1 = FactoryBot.create(:work, title: "表示される作品")
    work2 = FactoryBot.create(:work, title: "削除済み作品", deleted_at: Time.current)
    character = FactoryBot.create(:character)
    FactoryBot.create(:cast, person:, work: work1, character:)
    FactoryBot.create(:cast, person:, work: work2, character:)

    get "/people/#{person.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include("表示される作品")
    expect(response.body).not_to include("削除済み作品")
  end
end
