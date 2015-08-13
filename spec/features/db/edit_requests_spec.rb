require "spec_helper"

describe "Annict DB" do
  describe "編集リクエスト詳細ページ" do
    let(:user) { create(:registered_user, :with_editor_role) }
    let(:draft_work) { create(:draft_work) }
    let(:edit_request) { create(:edit_request, user: user, draft_resource: draft_work) }

    before do
      login_as(user, scope: :user)
      visit "/db/edit_requests/#{edit_request.id}"
    end

    it "ページが表示されること" do
      expect(page).to have_content(edit_request.title)
    end
  end
end
