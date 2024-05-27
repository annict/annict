# typed: false
# frozen_string_literal: true

describe "POST /db/trailers/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer, :unpublished) }

    it "user can not access this page" do
      post "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(trailer.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(trailer.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:trailer) { create(:trailer, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish trailer" do
      expect(trailer.published?).to eq(false)

      post "/db/trailers/#{trailer.id}/publishing"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(trailer.published?).to eq(true)
    end
  end
end
