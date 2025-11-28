# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/people", type: :request do
  it "ユーザーがサインインしていないとき、人物一覧が表示されること" do
    person = FactoryBot.create(:person)

    get "/db/people"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "ユーザーがサインインしているとき、人物一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    person = FactoryBot.create(:person)
    login_as(user, scope: :user)

    get "/db/people"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "ページネーションが動作すること" do
    # 100件以上の人物を作成してページネーションをテスト
    FactoryBot.create_list(:person, 101)

    get "/db/people", params: {page: 2}

    expect(response.status).to eq(200)
  end

  it "削除された人物は表示されないこと" do
    person = FactoryBot.create(:person)
    deleted_person = FactoryBot.create(:person, deleted_at: Time.current)

    get "/db/people"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
    expect(response.body).not_to include(deleted_person.name)
  end
end
