require "spec_helper"

describe "Annict DB" do
  describe "作品の編集リクエスト作成ページ" do
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
      visit "/db/draft_works/new"
    end

    it "ページが表示されること" do
      expect(page).to have_content("作品情報")
    end

    context "入力して送信したとき" do
      before do
        within("#new_draft_work") do
          fill_in "draft_work[title]", with: "のんのんびより"
          select "TV", from: "draft_work[media]"
          fill_in "draft_work[edit_request_attributes][title]", with: "「のんのんびよりを登録」"
          click_button "作成する"
        end
      end

      it "編集リクエストが登録されること" do
        edit_request = EditRequest.first
        expect(edit_request.title).to eq "「のんのんびよりを登録」"
        expect(edit_request.draft_resource.present?).to be true
      end

      it "編集リクエスト詳細ページに遷移すること" do
        edit_request = EditRequest.first
        expect(current_path).to eq "/db/edit_requests/#{edit_request.id}"
      end
    end
  end
end
