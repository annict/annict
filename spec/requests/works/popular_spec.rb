# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/popular", type: :request do
  it "ログインしていないとき、人気の作品一覧が表示されること" do
    work1 = FactoryBot.create(:work, watchers_count: 100)
    work2 = FactoryBot.create(:work, watchers_count: 200)
    work3 = FactoryBot.create(:work, watchers_count: 50)

    get "/works/popular"

    expect(response.status).to eq(200)
    expect(response.body).to include(work1.title)
    expect(response.body).to include(work2.title)
    expect(response.body).to include(work3.title)
    # watchers_countの降順で表示されることを確認
    expect(response.body.index(work2.title)).to be < response.body.index(work1.title)
    expect(response.body.index(work1.title)).to be < response.body.index(work3.title)
  end

  it "ログインしているとき、人気の作品一覧が表示されること" do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, watchers_count: 100)

    login_as(user, scope: :user)
    get "/works/popular"

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
  end

  it "削除された作品は表示されないこと" do
    kept_work = FactoryBot.create(:work, watchers_count: 100)
    deleted_work = FactoryBot.create(:work, :deleted, watchers_count: 200)

    get "/works/popular"

    expect(response.status).to eq(200)
    expect(response.body).to include(kept_work.title)
    expect(response.body).not_to include(deleted_work.title)
  end

  it "ページネーションが動作すること" do
    # 40件の作品を作成してページ2が存在することを確認
    works = []
    40.times do |i|
      works << FactoryBot.create(:work, watchers_count: 1000 - i)
    end

    # 1ページ目を確認
    get "/works/popular"
    expect(response.status).to eq(200)
    # 最初の30件が表示される
    expect(response.body).to include(works[0].title)
    expect(response.body).to include(works[29].title)
    expect(response.body).not_to include(works[30].title)

    # 2ページ目を確認
    get "/works/popular?page=2"
    expect(response.status).to eq(200)
    # 31件目から40件目が表示される
    expect(response.body).to include(works[30].title)
    expect(response.body).to include(works[39].title)
  end
end
