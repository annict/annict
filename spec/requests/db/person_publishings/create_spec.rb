# typed: false
# frozen_string_literal: true

describe "POST /db/people/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person, :unpublished) }

    it "user can not access this page" do
      post "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(person.published?).to eq(false)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:person) { create(:person, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(person.published?).to eq(false)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:person) { create(:person, :unpublished) }

    before do
      login_as(user, scope: :user)
    end

    it "user can publish person" do
      expect(person.published?).to eq(false)

      post "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("公開しました")

      expect(person.published?).to eq(true)
    end
  end
end
