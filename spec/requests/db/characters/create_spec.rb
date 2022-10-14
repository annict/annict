# frozen_string_literal: true

describe "POST /db/characters", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series) }
    let!(:character_params) do
      {
        rows: "かぐや姫,かぐやひめ,#{series.name}"
      }
    end

    it "user can not access this page" do
      post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Character.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:series) { create(:series) }
    let!(:user) { create(:registered_user) }
    let!(:character_params) do
      {
        rows: "かぐや姫,かぐやひめ,#{series.name}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Character.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:series) { create(:series) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:character_params) do
      {
        rows: "かぐや姫,かぐやひめ,#{series.name}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create character" do
      expect(Character.all.size).to eq(0)

      post "/db/characters", params: {deprecated_db_character_rows_form: character_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Character.all.size).to eq(1)
      character = Character.first

      expect(character.name).to eq("かぐや姫")
      expect(character.name_kana).to eq("かぐやひめ")
      expect(character.series_id).to eq(series.id)
    end
  end
end
