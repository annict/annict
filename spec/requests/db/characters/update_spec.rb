# typed: false
# frozen_string_literal: true

describe "PATCH /db/characters/:id", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character) }
    let!(:old_character) { character.attributes }
    let!(:character_params) do
      {
        name: "かぐや姫"
      }
    end

    it "user can not access this page" do
      patch "/db/characters/#{character.id}", params: {character: character_params}
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(character.name).to eq(old_character["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:character) { create(:character) }
    let!(:old_character) { character.attributes }
    let!(:character_params) do
      {
        name: "かぐや姫"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/characters/#{character.id}", params: {character: character_params}
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(character.name).to eq(old_character["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:character) { create(:character) }
    let!(:old_character) { character.attributes }
    let!(:character_params) do
      {
        name: "かぐや姫"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update character" do
      expect(character.name).to eq(old_character["name"])

      patch "/db/characters/#{character.id}", params: {character: character_params}
      character.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(character.name).to eq("かぐや姫")
    end
  end
end
