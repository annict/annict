# frozen_string_literal: true

describe "PATCH /db/people/:id", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person) }
    let!(:old_person) { person.attributes }
    let!(:person_params) do
      {
        name: "徳川家康"
      }
    end

    it "user can not access this page" do
      patch "/db/people/#{person.id}", params: {person: person_params}
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(person.name).to eq(old_person["name"])
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:person) { create(:person) }
    let!(:old_person) { person.attributes }
    let!(:person_params) do
      {
        name: "徳川家康"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/people/#{person.id}", params: {person: person_params}
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(person.name).to eq(old_person["name"])
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:person) { create(:person) }
    let!(:old_person) { person.attributes }
    let!(:person_params) do
      {
        name: "徳川家康"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can update person" do
      expect(person.name).to eq(old_person["name"])

      patch "/db/people/#{person.id}", params: {person: person_params}
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(person.name).to eq("徳川家康")
    end
  end
end
