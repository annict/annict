# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/search", type: :request do
  it "ログインしていないとき、リソースが保存されている場合、検索結果を表示すること" do
    series = create(:series, name: "しりーず検索")
    work = create(:work, title: "さくひん検索")
    person = create(:person, name: "じんぶつ検索")
    organization = create(:organization, name: "だんたい検索")
    character = create(:character, name: "きゃらくたー検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(person.name)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(character.name)
  end

  it "ログインしていないとき、リソースが保存されていない場合、登録されていませんと表示すること" do
    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "ログインしているとき、リソースが保存されている場合、検索結果を表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    series = create(:series, name: "しりーず検索")
    work = create(:work, title: "さくひん検索")
    person = create(:person, name: "じんぶつ検索")
    organization = create(:organization, name: "だんたい検索")
    character = create(:character, name: "きゃらくたー検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(series.name)
    expect(response.body).to include(work.title)
    expect(response.body).to include(person.name)
    expect(response.body).to include(organization.name)
    expect(response.body).to include(character.name)
  end

  it "ログインしているとき、リソースが保存されていない場合、登録されていませんと表示すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "管理者としてログインしているとき、公開中の作品に編集・非公開・削除のリンクを Go 版の URL で表示すること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(%(href="/db/works/#{work.id}/edit"))
    expect(response.body).to include(%(href="/db/works/#{work.id}/archive/new?return_to=))
    expect(response.body).to include(%(href="/db/works/#{work.id}/deletion/new?return_to=))
    expect(response.body).not_to include("/db/works/#{work.id}/unarchive/new")
  end

  it "管理者としてログインしているとき、非公開の作品には非公開ではなく公開のリンクを表示すること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索", unpublished_at: Time.zone.now)

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(%(href="/db/works/#{work.id}/unarchive/new?return_to=))
    expect(response.body).to include(%(href="/db/works/#{work.id}/deletion/new?return_to=))
    expect(response.body).not_to include("/db/works/#{work.id}/archive/new")
  end

  it "編集権限はあるが管理者ではないユーザーとしてログインしているとき、削除のリンクを表示しないこと" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(%(href="/db/works/#{work.id}/edit"))
    expect(response.body).to include(%(href="/db/works/#{work.id}/archive/new?return_to=))
    expect(response.body).not_to include("/db/works/#{work.id}/deletion/new")
  end

  it "編集権限が無いユーザーとしてログインしているとき、作品の操作リンクを表示しないこと" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    expect(response.body).to include(work.title)
    expect(response.body).not_to include(%(href="/db/works/#{work.id}/edit"))
    expect(response.body).not_to include("/db/works/#{work.id}/archive/new")
    expect(response.body).not_to include("/db/works/#{work.id}/deletion/new")
  end

  it "空の検索クエリの場合、正常にレスポンスを返すこと" do
    get "/db/search", params: {q: ""}

    expect(response.status).to eq(200)
  end

  it "検索クエリがnilの場合、正常にレスポンスを返すこと" do
    get "/db/search"

    expect(response.status).to eq(200)
  end

  it "ひらがな・カタカナの変換を考慮して検索結果を表示すること" do
    create(:series, name: "テストシリーズ")

    get "/db/search", params: {q: "てすと"}

    expect(response.status).to eq(200)
    expect(response.body).to include("登録されていません")
  end

  it "管理者としてログインしているとき、削除済み作品には公開状態変更・削除のリンクを表示しないこと" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    now = Time.zone.now
    works = [
      create(:work, title: "削除済み公開中作品", deleted_at: now),
      create(:work, title: "削除済み非公開作品", unpublished_at: now, deleted_at: now)
    ]

    get "/db/search", params: {q: "削除済み"}

    expect(response.status).to eq(200)
    works.each do |work|
      expect(response.body).to include(work.title)
      expect(response.body).not_to include("/db/works/#{work.id}/archive/new")
      expect(response.body).not_to include("/db/works/#{work.id}/unarchive/new")
      expect(response.body).not_to include("/db/works/#{work.id}/deletion/new")
    end
  end

  it "操作リンクのアクセシブルネームに、対象の作品を視覚的に隠したテキストとして含めること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    label = %(<span class="visually-hidden">#{I18n.t("messages.db.works.action_target_label", id: work.id)}</span>)
    expect(response.body.scan(label).size).to eq(3)
  end

  it "確認画面へのリンクが、検索結果のパスを return_to として渡すこと" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    work = create(:work, title: "さくひん検索")

    get "/db/search", params: {q: "検索"}

    expect(response.status).to eq(200)
    return_to = {return_to: request.fullpath}.to_query
    expect(response.body).to include(%(href="/db/works/#{work.id}/archive/new?#{return_to}"))
    expect(response.body).to include(%(href="/db/works/#{work.id}/deletion/new?#{return_to}"))
  end
end
