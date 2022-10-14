# frozen_string_literal: true

describe "POST /db/people", type: :request do
  context "user does not sign in" do
    let!(:person_params) do
      {
        rows: "徳川家康,とくがわいえやす"
      }
    end

    it "user can not access this page" do
      post "/db/people", params: {deprecated_db_person_rows_form: person_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Person.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:person_params) do
      {
        rows: "徳川家康,とくがわいえやす"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/people", params: {deprecated_db_person_rows_form: person_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Person.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:person_params) do
      {
        rows: "徳川家康,とくがわいえやす"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create person" do
      expect(Person.all.size).to eq(0)

      post "/db/people", params: {deprecated_db_person_rows_form: person_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Person.all.size).to eq(1)
      person = Person.first

      expect(person.name).to eq("徳川家康")
      expect(person.name_kana).to eq("とくがわいえやす")
    end
  end
end
