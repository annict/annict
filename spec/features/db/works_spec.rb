require "spec_helper"

describe "Annict DB" do
  describe "作品一覧ページ" do
    context "全て" do
      let!(:work) { create(:work) }

      before do
        visit "/db/works"
      end

      it "作品が表示されること" do
        expect(page).to have_content(work.title)
      end
    end

    context "今期" do
      let!(:work) { create(:work, :with_current_season) }

      before do
        visit "/db/works/season?slug=#{ENV['ANNICT_CURRENT_SEASON']}"
      end

      it "作品が表示されること" do
        expect(page).to have_content(work.title)
      end
    end

    context "来期" do
      let!(:work) { create(:work, :with_next_season) }

      before do
        visit "/db/works/season?slug=#{ENV['ANNICT_NEXT_SEASON']}"
      end

      it "作品が表示されること" do
        expect(page).to have_content(work.title)
      end
    end

    context "前期" do
      let!(:work) { create(:work, :with_prev_season) }

      before do
        visit "/db/works/season?slug=#{ENV['ANNICT_PREV_SEASON']}"
      end

      it "作品が表示されること" do
        expect(page).to have_content(work.title)
      end
    end

    context "エピソード未登録" do
      let!(:work1) { create(:work, :with_episode) }
      let!(:work2) { create(:work) }

      before do
        visit "/db/works/resourceless?name=episode"
      end

      it "エピソードが登録されていない作品が表示されること" do
        expect(page).to_not have_content(work1.title)
        expect(page).to have_content(work2.title)
      end
    end

    context "作品画像未登録" do
      let!(:work1) { create(:work, :with_item) }
      let!(:work2) { create(:work) }

      before do
        visit "/db/works/resourceless?name=item"
      end

      it "作品画像が登録されていない作品が表示されること" do
        expect(page).to_not have_content(work1.title)
        expect(page).to have_content(work2.title)
      end
    end
  end

  describe "作品作成ページ" do
    let(:user) { create(:registered_user, :with_editor_role) }

    before do
      login_as(user, scope: :user)
      visit "/db/works/new"
    end

    it "ページが表示されること" do
      expect(page).to have_content("リリース時期")
    end

    context "入力して送信したとき" do
      before do
        within("#new_work") do
          fill_in "work[title]", with: "ご注文はうさぎですか?"
          select "TV", from: "work[media]"
          click_button "登録する"
        end
      end

      it "編集リクエストが登録されること" do
        work = Work.find_by(title: "ご注文はうさぎですか?")
        expect(work.present?).to be true
      end

      it "作品編集ページに遷移すること" do
        work = Work.find_by(title: "ご注文はうさぎですか?")
        expect(current_path).to eq "/db/works/#{work.id}/edit"
      end
    end
  end
end
