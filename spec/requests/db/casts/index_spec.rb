# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/casts", type: :request do
  it "ユーザーがログインしていないとき、キャスト一覧が表示されること" do
    cast = create(:cast)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).to include(cast.person.name)
  end

  it "ユーザーがログインしているとき、キャスト一覧が表示されること" do
    user = create(:registered_user)
    cast = create(:cast)
    login_as(user, scope: :user)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).to include(cast.person.name)
  end

  it "削除済みのキャストは表示されないこと" do
    cast = create(:cast)
    deleted_cast = create(:cast, work: cast.work, deleted_at: Time.current)

    get "/db/works/#{cast.work_id}/casts"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
    expect(response.body).not_to include(deleted_cast.character.name)
  end

  it "キャストがソート番号順に表示されること" do
    work = create(:work)
    cast1 = create(:cast, work: work, sort_number: 2)
    cast2 = create(:cast, work: work, sort_number: 1)
    cast3 = create(:cast, work: work, sort_number: 3)

    get "/db/works/#{work.id}/casts"

    expect(response.status).to eq(200)
    # sort_number順で表示されているか確認
    body_index1 = response.body.index(cast1.character.name)
    body_index2 = response.body.index(cast2.character.name)
    body_index3 = response.body.index(cast3.character.name)
    expect(body_index2).to be < body_index1
    expect(body_index1).to be < body_index3
  end

  it "削除済みの作品にアクセスしたとき、RecordNotFoundエラーが発生すること" do
    work = create(:work, deleted_at: Time.current)

    expect {
      get "/db/works/#{work.id}/casts"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しない作品IDを指定したとき、RecordNotFoundエラーが発生すること" do
    expect {
      get "/db/works/999999/casts"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
