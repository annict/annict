# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:work_id/videos", type: :request do
  it "ユーザーがログインしていないとき、動画一覧が表示されること" do
    work = create(:work)
    trailer = create(:trailer, work:, title: "予告編1", sort_number: 10)

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "ユーザーがログインしているとき、動画一覧が表示されること" do
    user = create(:registered_user)
    work = create(:work)
    trailer = create(:trailer, work:, title: "予告編1")
    login_as(user, scope: :user)

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    expect(response.body).to include(trailer.title)
  end

  it "複数の動画があるとき、sort_number順に表示されること" do
    work = create(:work)
    trailer1 = create(:trailer, work:, title: "動画2", sort_number: 20)
    trailer2 = create(:trailer, work:, title: "動画1", sort_number: 10)
    trailer3 = create(:trailer, work:, title: "動画3", sort_number: 30)

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    # sort_number順（10, 20, 30）で表示されるか確認
    body = response.body
    trailer2_pos = body.index(trailer2.title)
    trailer1_pos = body.index(trailer1.title)
    trailer3_pos = body.index(trailer3.title)
    expect(trailer2_pos).to be < trailer1_pos
    expect(trailer1_pos).to be < trailer3_pos
  end

  it "動画が存在しないとき、空の一覧が表示されること" do
    work = create(:work)

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    expect(response.body).to include(I18n.t("messages._empty.no_resources"))
  end

  it "削除された動画は表示されないこと" do
    work = create(:work)
    active_trailer = create(:trailer, work:, title: "アクティブな動画")
    deleted_trailer = create(:trailer, :deleted, work:, title: "削除された動画")

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    expect(response.body).to include(active_trailer.title)
    expect(response.body).not_to include(deleted_trailer.title)
  end

  it "unpublishedの動画は表示されないこと" do
    work = create(:work)
    published_trailer = create(:trailer, work:, title: "公開中の動画")
    unpublished_trailer = create(:trailer, :unpublished, work:, title: "非公開の動画")

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(200)
    expect(response.body).to include(published_trailer.title)
    expect(response.body).not_to include(unpublished_trailer.title)
  end

  it "存在しない作品IDが指定されたとき、404エラーが返されること" do
    get "/works/99999999/videos"

    expect(response.status).to eq(404)
  end

  it "削除された作品の動画一覧にアクセスしたとき、404エラーが返されること" do
    work = create(:work, :deleted)

    get "/works/#{work.id}/videos"

    expect(response.status).to eq(404)
  end
end
