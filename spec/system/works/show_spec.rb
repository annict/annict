# typed: false
# frozen_string_literal: true

RSpec.describe "Works#show page", type: :system do
  it "作品詳細ページが表示されること" do
    work = FactoryBot.create(:work, :with_current_season)

    visit work_path(work)

    expect(page).to have_content(work.title)
    expect(page).to have_http_status(:ok)
  end

  it "ログイン済みユーザーがステータスを選択できること", js: true do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :with_current_season)

    sign_in(user:)
    visit work_path(work)

    # JavaScriptの読み込みを待つ
    wait_for_javascript

    # ドロップダウンボタンを探してクリック
    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")
    dropdown_button.click

    # ドロップダウンメニューが表示されることを確認
    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    # 「見たい」ステータスを選択
    within(".c-status-select-dropdown .dropdown-menu") do
      find("button", text: "見たい").click
    end

    # APIリクエストが完了するまで待つ
    sleep 2

    # ボタンのクラスが更新されることを確認（アイコンのみ表示される）
    # plan_to_watchのアイコンはcircle
    expect(dropdown_button).to have_css(".fa-circle")
    expect(dropdown_button).to have_css(".u-bg-plan-to-watch")
  end

  it "ログイン済みユーザーが異なるステータスを選択できること", js: true do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :with_current_season)

    sign_in(user:)
    visit work_path(work)

    wait_for_javascript

    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")

    # 各ステータスをテスト
    statuses = ["見たい", "見てる", "見た", "一時中断", "視聴中止"]

    statuses.each do |status|
      dropdown_button.click
      dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
      expect(dropdown_menu).to be_visible

      within(".c-status-select-dropdown .dropdown-menu") do
        find("button", text: status).click
      end

      # APIリクエストが完了するまで待つ
      sleep 2

      # ボタンのクラスが更新されることを確認
      case status
      when "見たい"
        expect(dropdown_button).to have_css(".fa-circle")
        expect(dropdown_button).to have_css(".u-bg-plan-to-watch")
      when "見てる"
        expect(dropdown_button).to have_css(".fa-play")
        expect(dropdown_button).to have_css(".u-bg-watching")
      when "見た"
        expect(dropdown_button).to have_css(".fa-check")
        expect(dropdown_button).to have_css(".u-bg-completed")
      when "一時中断"
        expect(dropdown_button).to have_css(".fa-pause")
        expect(dropdown_button).to have_css(".u-bg-on-hold")
      when "視聴中止"
        expect(dropdown_button).to have_css(".fa-stop")
        expect(dropdown_button).to have_css(".u-bg-dropped")
      end

    end
  end

  it "ログイン済みユーザーが「ステータスを外す」を選択できること", js: true do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :with_current_season)
    # 既にステータスが設定されている状態を作成
    FactoryBot.create(:status, user:, work:, kind: :watching)

    sign_in(user:)
    visit work_path(work)

    wait_for_javascript

    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")

    # APIリクエストが完了するまで待つ
    sleep 2

    # 初期状態で「見てる」アイコンが表示されていることを確認
    # watchingのアイコンはplay
    expect(dropdown_button).to have_css(".fa-play")
    expect(dropdown_button).to have_css(".u-bg-watching")

    dropdown_button.click
    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    within(".c-status-select-dropdown .dropdown-menu") do
      find("button", text: "ステータスを外す").click
    end

    # APIリクエストが完了するまで待つ
    sleep 2

    # ボタンがデフォルト状態に戻ることを確認
    # no_statusのアイコンはbars
    expect(dropdown_button).to have_css(".fa-bars")
    expect(dropdown_button).not_to have_css(".u-bg-watching")
  end

  it "未ログインユーザーがステータスを選択しようとするとサインアップモーダルが表示されること", js: true do
    work = FactoryBot.create(:work, :with_current_season)

    visit work_path(work)

    wait_for_javascript

    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")
    dropdown_button.click

    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    within(".c-status-select-dropdown .dropdown-menu") do
      find("button", text: "見たい").click
    end

    # サインアップモーダルが表示されることを確認
    expect(page).to have_css(".c-sign-up-modal", visible: true)
    within(".c-sign-up-modal") do
      expect(page).to have_content("アカウント作成")
    end
  end

  it "エピソード一覧が表示されること" do
    work = FactoryBot.create(:work, :with_current_season)
    episodes = FactoryBot.create_list(:episode, 5, work:)

    visit work_path(work)

    # エピソードセクションが存在することを確認
    expect(page).to have_content("エピソード")

    # 各エピソードが表示されることを確認
    episodes.each do |episode|
      expect(page).to have_content(episode.number)
      expect(page).to have_content(episode.title) if episode.title.present?
    end
  end

  it "コメント一覧が表示されること" do
    work = FactoryBot.create(:work, :with_current_season)
    users = FactoryBot.create_list(:registered_user, 3)

    records = users.map do |user|
      record = FactoryBot.create(:record, user:, work:)
      FactoryBot.create(:work_record, record:, user:, work:, body: "#{user.username}のコメント")
      record
    end

    visit work_path(work)

    # コメントセクションが存在することを確認
    expect(page).to have_content("感想")

    # 各コメントが表示されることを確認
    records.each do |record|
      expect(page).to have_content(record.work_record.body)
      expect(page).to have_content(record.user.username)
    end
  end

  it "あらすじが表示されること" do
    synopsis = "これはテスト用のあらすじです。"
    work = FactoryBot.create(:work, :with_current_season, synopsis:)

    visit work_path(work)

    expect(page).to have_content("あらすじ")
    expect(page).to have_content(synopsis)
  end

  it "動画がある場合は動画セクションが表示されること" do
    work = FactoryBot.create(:work, :with_current_season)
    trailer = FactoryBot.create(:trailer, work:, title: "予告編第1弾")

    visit work_path(work)

    expect(page).to have_content("動画")
    expect(page).to have_content(trailer.title)
    expect(page).to have_css(".c-video-thumbnail")
  end
end
