# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/trailers", type: :request do
  it "ユーザーがログインしていないとき、トレイラー一覧が表示されること" do
    trailer = create(:trailer)

    get "/db/works/#{trailer.work_id}/trailers"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "ユーザーがログインしているとき、トレイラー一覧が表示されること" do
    user = create(:registered_user)
    trailer = create(:trailer)
    login_as(user, scope: :user)

    get "/db/works/#{trailer.work_id}/trailers"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "複数のトレイラーがあるとき、sort_number順に表示されること" do
    work = create(:work)
    trailer1 = create(:trailer, work: work, title: "トレイラー2", sort_number: 20)
    trailer2 = create(:trailer, work: work, title: "トレイラー1", sort_number: 10)
    trailer3 = create(:trailer, work: work, title: "トレイラー3", sort_number: 30)

    get "/db/works/#{work.id}/trailers"

    expect(response.status).to eq(200)
    # sort_number順（10, 20, 30）で表示されるか確認
    body = response.body
    trailer2_pos = body.index(trailer2.title)
    trailer1_pos = body.index(trailer1.title)
    trailer3_pos = body.index(trailer3.title)
    expect(trailer2_pos).to be < trailer1_pos
    expect(trailer1_pos).to be < trailer3_pos
  end

  it "トレイラーが存在しないとき、空の一覧が表示されること" do
    work = create(:work)

    get "/db/works/#{work.id}/trailers"

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "削除されたトレイラーは表示されないこと" do
    work = create(:work)
    active_trailer = create(:trailer, work: work, title: "アクティブなトレイラー")
    deleted_trailer = create(:trailer, work: work, title: "削除されたトレイラー", deleted_at: Time.current)

    get "/db/works/#{work.id}/trailers"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_trailer.title)
    expect(response.body).not_to include(deleted_trailer.title)
  end
end
