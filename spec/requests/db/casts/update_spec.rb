# frozen_string_literal: true

describe "PATCH /db/casts/:id", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:cast) { create(:cast) }
    let!(:old_cast) { cast.attributes }
    let!(:cast_params) do
      {
        character_id: character.id,
        person_id: person.id
      }
    end

    it "user can not access this page" do
      patch "/db/casts/#{cast.id}", params: {cast: cast_params}
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(cast.character_id).to eq(old_cast["character_id"])
      expect(cast.person_id).to eq(old_cast["person_id"])
    end
  end

  context "user who is not editor signs in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:user) { create(:registered_user) }
    let!(:cast) { create(:cast) }
    let!(:old_cast) { cast.attributes }
    let!(:cast_params) do
      {
        character_id: character.id,
        person_id: person.id
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/casts/#{cast.id}", params: {cast: cast_params}
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(cast.character_id).to eq(old_cast["character_id"])
      expect(cast.person_id).to eq(old_cast["person_id"])
    end
  end

  context "user who is editor signs in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:cast) { create(:cast) }
    let!(:old_cast) { cast.attributes }
    let!(:cast_params) do
      {
        character_id: character.id,
        person_id: person.id
      }
    end
    let!(:attr_names) { cast_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update cast" do
      expect(cast.character_id).to eq(old_cast["character_id"])
      expect(cast.person_id).to eq(old_cast["person_id"])

      patch "/db/casts/#{cast.id}", params: {cast: cast_params}
      cast.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(cast.character_id).to eq(character.id)
      expect(cast.person_id).to eq(person.id)
    end
  end
end
