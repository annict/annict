# typed: false
# frozen_string_literal: true

describe "DELETE /db/trailers/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer, :published) }

    it "user can not access this page" do
      delete "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(trailer.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(trailer.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:trailer) { create(:trailer, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish trailer" do
      expect(trailer.published?).to eq(true)

      delete "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(trailer.published?).to eq(false)
    end
  end
end
