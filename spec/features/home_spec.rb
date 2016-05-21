# frozen_string_literal: true

describe "トップページ" do
  context "ログインしていないとき" do
    let!(:work) { create(:work, :with_item, :with_current_season) }

    before do
      visit "/"
    end

    it "アクセスするとページが表示される" do
      expect(page).to have_content("見たアニメを記録して、共有しよう")
    end

    it "Twitterでログインをクリックするとユーザ登録ページが表示される" do
      find(".ann-navbar .sign-up").click
      find("#signup-modal").click_link("Twitterアカウントで登録")

      expect(current_path).to eq "/sign_up"
    end

    it "Facebookでログインをクリックするとユーザ登録ページが表示される" do
      find(".ann-navbar .sign-up").click
      find("#signup-modal").click_link("Facebookアカウントで登録")

      expect(current_path).to eq "/sign_up"
    end
  end

  context "ログインしているとき" do
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "アクティビティが存在しないとき" do
      before do
        visit "/"
      end

      it "アクティビティが存在しない旨を表示すること", js: true do
        expect(page).to have_content("アクティビティはありません")
      end
    end

    context "自分が記録しているとき" do
      let!(:record_tip) { create(:record_tip) }
      let(:episode) { create(:episode) }

      before do
        Delayed::Worker.delay_jobs = false

        record = user.checkins.new do |c|
          c.work = episode.work
          c.episode = episode
          c.comment = "おもしろかったよ"
          c.rating = 3.0
        end
        NewRecordService.new(user, record).save

        visit "/"
      end

      it "アクティビティに自分の記録情報が表示されること", js: true do
        expect(page).to have_content("おもしろかったよ")
      end
    end

    context "フォローしている人が記録しているとき" do
      let!(:record_tip) { create(:record_tip) }
      let!(:following_user) { create(:registered_user) }
      let(:episode) { create(:episode) }

      before do
        Delayed::Worker.delay_jobs = false

        record = following_user.checkins.new do |c|
          c.work = episode.work
          c.episode = episode
          c.comment = "たのしかったよ"
          c.rating = 3.0
        end
        NewRecordService.new(user, record).save

        user.follow(following_user)

        visit "/"
      end

      it "アクティビティにフォローしている人の記録情報が表示されること", js: true do
        expect(page).to have_content("たのしかったよ")
      end
    end
  end
end
