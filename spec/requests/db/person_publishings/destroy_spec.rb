# typed: false
# frozen_string_literal: true

describe "DELETE /db/people/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:person) { create(:person, :published) }

    it "user can not access this page" do
      delete "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(person.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:person) { create(:person, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(person.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:person) { create(:person, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish person" do
      expect(person.published?).to eq(true)

      delete "/db/people/#{person.id}/publishing"
      person.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(person.published?).to eq(false)
    end
  end
end
