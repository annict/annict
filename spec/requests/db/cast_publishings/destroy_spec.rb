# typed: false
# frozen_string_literal: true

describe "DELETE /db/casts/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:cast) { create(:cast, :published) }

    it "user can not access this page" do
      delete "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(cast.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:cast) { create(:cast, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(cast.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:cast) { create(:cast, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish cast" do
      expect(cast.published?).to eq(true)

      delete "/db/casts/#{cast.id}/publishing"
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(cast.published?).to eq(false)
    end
  end
end
