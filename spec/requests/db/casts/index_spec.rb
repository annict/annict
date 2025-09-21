# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/casts", type: :request do
  it "ユーザーがログインしていないとき、キャスト一覧が表示されること" do
    cast = FactoryBot.create(:cast)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).to include(cast.person.name)
  end

  it "ユーザーがログインしているとき、キャスト一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    cast = FactoryBot.create(:cast)
    login_as(user, scope: :user)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).to include(cast.person.name)
  end

  it "削除済みのキャストは表示されないこと" do
    cast = FactoryBot.create(:cast)
    deleted_cast = FactoryBot.create(:cast, work: cast.work, deleted_at: Time.current)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).not_to include(deleted_cast.character.name)
  end

  it "キャストがソート番号順に表示されること" do
    work = FactoryBot.create(:work)
    cast1 = FactoryBot.create(:cast, work:, sort_number: 2)
    cast2 = FactoryBot.create(:cast, work:, sort_number: 1)
    cast3 = FactoryBot.create(:cast, work:, sort_number: 3)

    get "/db/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    # sort_number順で表示されているか確認
    body_index1 = response.body.index(cast1.character.name)
    body_index2 = response.body.index(cast2.character.name)
    body_index3 = response.body.index(cast3.character.name)
    expect(body_index2).to be < body_index1
    expect(body_index1).to be < body_index3
  end

  it "削除済みの作品にアクセスしたとき、404エラーが返されること" do
    work = FactoryBot.create(:work, deleted_at: Time.current)

    expect {
      get "/db/works/#{work.id}/casts"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない作品IDを指定したとき、404エラーが返されること" do
    get "/db/works/999999/casts"

    expect(response.status).to eq(404)
  end

  it "キャストが存在しない作品にアクセスしたとき、空のリストが表示されること" do
    work = FactoryBot.create(:work)

    get "/db/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include("キャスト")
  end

  it "複数の削除済みキャストと通常のキャストが混在するとき、削除済みキャストは表示されないこと" do
    work = FactoryBot.create(:work)
    cast1 = FactoryBot.create(:cast, work:, sort_number: 1)
    deleted_cast1 = FactoryBot.create(:cast, work:, sort_number: 2, deleted_at: Time.current)
    cast2 = FactoryBot.create(:cast, work:, sort_number: 3)
    deleted_cast2 = FactoryBot.create(:cast, work:, sort_number: 4, deleted_at: Time.current)
    cast3 = FactoryBot.create(:cast, work:, sort_number: 5)

    get "/db/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    # 通常のキャストは表示される
    expect(response.body).to include(cast1.character.name)
    expect(response.body).to include(cast2.character.name)
    expect(response.body).to include(cast3.character.name)
    # 削除済みキャストは表示されない
    expect(response.body).not_to include(deleted_cast1.character.name)
    expect(response.body).not_to include(deleted_cast2.character.name)
  end

  it "キャラクター名やパーソン名にHTMLエスケープが必要な文字が含まれる場合、適切にエスケープされること" do
    work = FactoryBot.create(:work)
    character = FactoryBot.create(:character, name: '<script>alert("XSS")</script>')
    person = FactoryBot.create(:person, name: '"テスト" & テスト')
    FactoryBot.create(:cast, work:, character:, person:)

    get "/db/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    # HTMLエスケープされた文字列が含まれていることを確認
    expect(response.body).to include("&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;")
    expect(response.body).to include("&quot;テスト&quot; &amp; テスト")
  end
end
