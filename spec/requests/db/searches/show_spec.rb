# typed: false
# frozen_string_literal: true

describe "GET /db/search", type: :request do
  context "user does not sign in" do
    context "when resources are saved" do
      let!(:series) { create(:series, name: "しりーず検索") }
      let!(:work) { create(:work, title: "さくひん検索") }
      let!(:person) { create(:person, name: "じんぶつ検索") }
      let!(:organization) { create(:organization, name: "だんたい検索") }
      let!(:character) { create(:character, name: "きゃらくたー検索") }

      it "responses search result" do
        get "/db/search", params: {q: "検索"}

        expect(response.status).to eq(200)
        expect(response.body).to include(series.name)
        expect(response.body).to include(work.title)
        expect(response.body).to include(person.name)
        expect(response.body).to include(organization.name)
        expect(response.body).to include(character.name)
      end
    end

    context "when resources are not saved" do
      it "responses search result" do
        get "/db/search", params: {q: "検索"}

        expect(response.status).to eq(200)
        expect(response.body).to include("登録されていません")
      end
    end
  end

  context "user signs in" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context "when resources are saved" do
      let!(:series) { create(:series, name: "しりーず検索") }
      let!(:work) { create(:work, title: "さくひん検索") }
      let!(:person) { create(:person, name: "じんぶつ検索") }
      let!(:organization) { create(:organization, name: "だんたい検索") }
      let!(:character) { create(:character, name: "きゃらくたー検索") }

      it "responses search result" do
        get "/db/search", params: {q: "検索"}

        expect(response.status).to eq(200)
        expect(response.body).to include(series.name)
        expect(response.body).to include(work.title)
        expect(response.body).to include(person.name)
        expect(response.body).to include(organization.name)
        expect(response.body).to include(character.name)
      end
    end

    context "when resources are not saved" do
      it "responses search result" do
        get "/db/search", params: {q: "検索"}

        expect(response.status).to eq(200)
        expect(response.body).to include("登録されていません")
      end
    end
  end
end
