# typed: false
# frozen_string_literal: true

RSpec.describe "GET /works/:season_slug", type: :request do
  it "ユーザーがログインしていないとき、指定されたシーズンの作品一覧が表示されること" do
    spring_work = create(:work, season_year: 2023, season_name: "spring", watchers_count: 100)
    summer_work = create(:work, season_year: 2023, season_name: "summer", watchers_count: 50)
    # 削除された作品は表示されない
    deleted_work = create(:work, :deleted, season_year: 2023, season_name: "spring")

    get "/works/2023-spring"

    expect(response.status).to eq(200)
    expect(response.body).to include(spring_work.title)
    expect(response.body).not_to include(summer_work.title)
    expect(response.body).not_to include(deleted_work.title)
  end

  it "ユーザーがログインしているとき、指定されたシーズンの作品一覧が表示されること" do
    user = create(:registered_user)
    spring_work = create(:work, season_year: 2023, season_name: "spring")
    summer_work = create(:work, season_year: 2023, season_name: "summer")
    login_as(user, scope: :user)

    get "/works/2023-spring"

    expect(response.status).to eq(200)
    expect(response.body).to include(spring_work.title)
    expect(response.body).not_to include(summer_work.title)
  end

  it "all シーズンを指定したとき、その年の全ての作品が表示されること" do
    spring_work = create(:work, season_year: 2023, season_name: "spring")
    summer_work = create(:work, season_year: 2023, season_name: "summer")
    autumn_work = create(:work, season_year: 2023, season_name: "autumn")
    winter_work = create(:work, season_year: 2023, season_name: "winter")
    # 別の年の作品は表示されない
    other_year_work = create(:work, season_year: 2022, season_name: "spring")

    get "/works/2023-all"

    expect(response.status).to eq(200)
    expect(response.body).to include(spring_work.title)
    expect(response.body).to include(summer_work.title)
    expect(response.body).to include(autumn_work.title)
    expect(response.body).to include(winter_work.title)
    expect(response.body).not_to include(other_year_work.title)
  end

  it "視聴者数が多い順に作品が表示されること" do
    work1 = create(:work, season_year: 2023, season_name: "spring", watchers_count: 50, title: "作品1")
    work2 = create(:work, season_year: 2023, season_name: "spring", watchers_count: 100, title: "作品2")
    work3 = create(:work, season_year: 2023, season_name: "spring", watchers_count: 75, title: "作品3")

    get "/works/2023-spring"

    expect(response.status).to eq(200)
    # 視聴者数が多い順（100 > 75 > 50）
    body = response.body
    work2_index = body.index(work2.title)
    work3_index = body.index(work3.title)
    work1_index = body.index(work1.title)
    expect(work2_index).to be < work3_index
    expect(work3_index).to be < work1_index
  end

  it "ページネーションが機能すること" do
    # 31件の作品を作成（1ページ30件なので2ページ目が必要）
    31.times do |i|
      create(:work, season_year: 2023, season_name: "spring", title: "作品#{i + 1}", watchers_count: 31 - i)
    end

    # 1ページ目の確認
    get "/works/2023-spring"
    expect(response.status).to eq(200)
    expect(response.body).to include("作品1") # 視聴者数が最も多い
    expect(response.body).not_to include("作品31") # 視聴者数が最も少ない

    # 2ページ目の確認
    get "/works/2023-spring", params: {page: 2}
    expect(response.status).to eq(200)
    expect(response.body).to include("作品31")
    expect(response.body).not_to include("作品1")
  end

  it "grid表示モードのとき、1ページに30件表示されること" do
    35.times do |i|
      create(:work, season_year: 2023, season_name: "spring", title: "作品#{i + 1}", watchers_count: 35 - i)
    end

    get "/works/2023-spring", params: {display: "grid"}

    expect(response.status).to eq(200)
    # 1ページ目には30件表示される（視聴者数順）
    30.times do |i|
      expect(response.body).to include("作品#{i + 1}")
    end
    # 31件目以降は表示されない
    5.times do |i|
      expect(response.body).not_to include("作品#{i + 31}")
    end
  end

  it "grid_small表示モードのとき、1ページに120件表示されること" do
    125.times do |i|
      create(:work, season_year: 2023, season_name: "spring", title: "作品#{i + 1}", watchers_count: 125 - i)
    end

    get "/works/2023-spring", params: {display: "grid_small"}

    expect(response.status).to eq(200)
    # 1ページ目には120件表示される（視聴者数順）
    120.times do |i|
      expect(response.body).to include("作品#{i + 1}")
    end
    # 121件目以降は表示されない
    5.times do |i|
      expect(response.body).not_to include("作品#{i + 121}")
    end
  end

  it "無効なdisplayパラメータが指定されたとき、デフォルトのgrid表示になること" do
    35.times do |i|
      create(:work, season_year: 2023, season_name: "spring", title: "作品#{i + 1}", watchers_count: 35 - i)
    end

    get "/works/2023-spring", params: {display: "invalid"}

    expect(response.status).to eq(200)
    # デフォルトのgrid表示（30件）になる（視聴者数順）
    30.times do |i|
      expect(response.body).to include("作品#{i + 1}")
    end
    5.times do |i|
      expect(response.body).not_to include("作品#{i + 31}")
    end
  end

  it "存在しないシーズンが指定されたとき、エラーになること" do
    # 9999年はYEAR_LISTに含まれないため、Season.find_by_slugがnilを返す
    expect {
      get "/works/9999-spring"
    }.to raise_error(NoMethodError)
  end
end
