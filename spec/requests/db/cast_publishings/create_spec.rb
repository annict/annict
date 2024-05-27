# typed: false
# frozen_string_literal: true

describe "POST /db/casts/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:cast) { create(:cast, :unpublished) }

    it "user can not access this page" do
      post "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(cast.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:cast) { create(:cast, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(cast.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:cast) { create(:cast, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish cast" do
      expect(cast.published?).to eq(false)

      post "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(cast.published?).to eq(true)
    end
  end
end
