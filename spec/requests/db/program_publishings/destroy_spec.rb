# typed: false
# frozen_string_literal: true

describe "DELETE /db/programs/:id/publishing", type: :request do
  context "user does not sign in" do
    let!(:program) { create(:program, :published) }

    it "user can not access this page" do
      delete "/db/programs/#{program.id}/publishing"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(program.published?).to eq(true)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:program) { create(:program, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      delete "/db/programs/#{program.id}/publishing"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(program.published?).to eq(true)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:program) { create(:program, :published) }

    before do
      login_as(user, scope: :user)
    end

    it "user can unpublish program" do
      expect(program.published?).to eq(true)

      delete "/db/programs/#{program.id}/publishing"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("非公開にしました")

      expect(program.published?).to eq(false)
    end
  end
end
