# typed: false
# frozen_string_literal: true

describe "DELETE /db/characters/:id", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character, :not_deleted) }

    it "user can not access this page" do
      expect(Character.count).to eq(1)

      delete "/db/characters/#{character.id}"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Character.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:character) { create(:character, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Character.count).to eq(1)

      delete "/db/characters/#{character.id}"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Character.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:character) { create(:character, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Character.count).to eq(1)

      delete "/db/characters/#{character.id}"
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Character.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:character) { create(:character, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete character softly" do
      expect(Character.count).to eq(1)

      delete "/db/characters/#{character.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Character.count).to eq(0)
    end
  end
end
