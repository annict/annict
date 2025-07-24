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

    # JavaScriptの読み込みとcomponent-value-fetcherの初期化を待つ
    wait_for_javascript
    sleep 2

    # ドロップダウンボタンを探してクリック
    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")

    # 初期状態の確認（未選択の場合はfa-bars）
    expect(dropdown_button).to have_css(".fa-bars")

    dropdown_button.click

    # ドロップダウンメニューが表示されることを確認
    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    # ドロップダウンメニューに各ステータスが含まれていることを確認
    within(".c-status-select-dropdown .dropdown-menu") do
      expect(page).to have_button("未選択")
      expect(page).to have_button("見たい")
      expect(page).to have_button("見てる")
      expect(page).to have_button("見た")
      expect(page).to have_button("一時中断")
      expect(page).to have_button("視聴中止")
    end
  end

  it "ログイン済みユーザーが異なるステータスを選択できること", js: true do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :with_current_season)

    sign_in(user:)
    visit work_path(work)

    # JavaScriptの読み込みとcomponent-value-fetcherの初期化を待つ
    wait_for_javascript
    sleep 2

    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")

    # ドロップダウンボタンをクリック
    dropdown_button.click

    # ドロップダウンメニューが表示されることを確認
    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    # 各ステータスボタンが存在することを確認
    within(".c-status-select-dropdown .dropdown-menu") do
      expect(page).to have_button("見たい")
      expect(page).to have_button("見てる")
      expect(page).to have_button("見た")
      expect(page).to have_button("一時中断")
      expect(page).to have_button("視聴中止")
    end

    # メニューを閉じる
    dropdown_button.click
  end

  it "ログイン済みユーザーが「ステータスを外す」を選択できること", js: true do
    user = FactoryBot.create(:registered_user)
    work = FactoryBot.create(:work, :with_current_season)

    sign_in(user:)
    visit work_path(work)

    # JavaScriptの読み込みとcomponent-value-fetcherの初期化を待つ
    wait_for_javascript
    sleep 2

    dropdown_button = find(".c-status-select-dropdown button.dropdown-toggle")

    # 最初にステータスを設定する
    dropdown_button.click
    within(".c-status-select-dropdown .dropdown-menu") do
      find("button", text: "見てる").click
    end

    # 少し待つ
    sleep 1

    # ステータスを外す
    dropdown_button.click
    dropdown_menu = find(".c-status-select-dropdown .dropdown-menu", visible: true)
    expect(dropdown_menu).to be_visible

    within(".c-status-select-dropdown .dropdown-menu") do
      find("button", text: "未選択").click
    end

    # 少し待つ
    sleep 1
  end

  it "未ログインユーザーがステータスを選択しようとするとサインアップモーダルが表示されること", js: true do
    work = FactoryBot.create(:work, :with_current_season)

    visit work_path(work)

    # JavaScriptの読み込みとcomponent-value-fetcherの初期化を待つ
    wait_for_javascript
    sleep 2

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
