# frozen_string_literal: true

describe "POST /db/works/:work_id/casts", type: :request do
  context "user does not sign in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:work) { create(:work) }
    let!(:form_params) do
      {
        rows: "#{character.id},#{person.id}"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/casts", params: {db_cast_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Cast.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "#{character.id},#{person.id}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/casts", params: {db_cast_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Cast.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:character) { create(:character) }
    let!(:person) { create(:person) }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "#{character.id},#{person.id}"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create cast" do
      expect(Cast.all.size).to eq(0)

      post "/db/works/#{work.id}/casts", params: {db_cast_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Cast.all.size).to eq(1)
      cast = Cast.last

      expect(cast.character_id).to eq(character.id)
      expect(cast.person_id).to eq(person.id)
    end
  end
end
