# typed: false
# frozen_string_literal: true

describe "POST /db/characters/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character, :unpublished) }

    it "user can not access this page" do
      post "/db/characters/#{character.id}/publishing"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(character.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:character) { create(:character, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/characters/#{character.id}/publishing"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(character.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:character) { create(:character, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish character" do
      expect(character.published?).to eq(false)

      post "/db/characters/#{character.id}/publishing"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(character.published?).to eq(true)
    end
  end
end
