# typed: false
# frozen_string_literal: true

describe "POST /db/works/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:work, :unpublished) }

    it "user can not access this page" do
      post "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(work.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:work, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(work.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:work) { create(:work, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish work" do
      expect(work.published?).to eq(false)

      post "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(work.published?).to eq(true)
    end
  end
end
