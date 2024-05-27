# typed: false
# frozen_string_literal: true

describe "GET /db/slots/:id/edit", type: :request do
  context "user does not sign in" do
    let!(:slot) { create(:slot) }

    it "user can not access this page" do
      get "/db/slots/#{slot.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:slot) { create(:slot) }

    before do
      login_as(user, scope: :user)
    end

    it "can not access" do
      get "/db/slots/#{slot.id}/edit"

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:slot) { create(:slot) }

    before do
      login_as(user, scope: :user)
    end

    it "responses slot edit form" do
      get "/db/slots/#{slot.id}/edit"

      expect(response.status).to eq(200)
      expect(response.body).to include(slot.channel.name)
    end
  end
end
