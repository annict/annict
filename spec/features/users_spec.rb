# frozen_string_literal: true

describe "プロフィールページ" do
  let!(:record_tip) { create(:record_tip) }
  let(:user) { create(:registered_user) }
  let(:work) { create(:work, :with_item) }
  let(:episode) { create(:episode, work: work) }

  before do
    visit "/@#{user.username}"
  end

  it "ページが表示されること" do
    expect(find(".profile h1")).to have_content(user.profile.name)
  end

  describe "アクティビティ" do
    before do
      Delayed::Worker.delay_jobs = false

      record = user.checkins.new do |c|
        c.work = episode.work
        c.episode = episode
        c.comment = "おもしろかったよ"
        c.rating = 3.0
      end
      NewRecordService.new(user, record).save

      visit "/@#{user.username}"
    end

    it "記録情報が表示されること", js: true do
      expect(find(".ann-activities")).to have_content("おもしろかったよ")
    end
  end

  describe "見てるアニメ" do
    let!(:status_tip) { create(:status_tip) }

    before do
      user.statuses.create(work: work, kind: :watching)

      visit "/@#{user.username}"
    end

    it "作品が表示されること" do
      expect(find(".watching-works")).to have_content(work.title)
    end
  end
end

describe "見てる作品一覧ページ" do
  let(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  context "見てる作品があるとき" do
    let!(:status_tip) { create(:status_tip) }
    let(:work) { create(:work, :with_item) }

    before do
      user.statuses.create(work: work, kind: :watching)

      visit "/@#{user.username}/watching"
    end

    it "見てる作品が表示されること" do
      expect(page).to have_content(work.title)
    end
  end
end

describe "アカウント設定ページ" do
  let(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
    visit "/settings/account"
  end

  context "メールアドレスを変更したとき" do
    before do
      within("form.edit_user") do
        fill_in "user_email", with: "fumoffu@example.com"
        click_button "更新する"
      end
    end

    it "要確認メールアドレスとしてデータベースに保存されること" do
      expect(user.reload.unconfirmed_email).to eq("fumoffu@example.com")
    end
  end
end
