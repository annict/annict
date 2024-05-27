# typed: false
# frozen_string_literal: true

describe "DELETE /db/works/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:work, :published) }

    it "user can not access this page" do
      delete "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(work.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:work, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(work.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:work) { create(:work, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish work" do
      expect(work.published?).to eq(true)

      delete "/db/works/#{work.id}/publishing"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(work.published?).to eq(false)
    end
  end
end
