# typed: false
# frozen_string_literal: true

describe "DELETE /db/programs/:id", type: :request do
  context "user does not sign in" do
    let!(:program) { create(:program, :not_deleted) }

    it "user can not access this page" do
      expect(Program.count).to eq(1)

      delete "/db/programs/#{program.id}"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Program.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:program) { create(:program, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Program.count).to eq(1)

      delete "/db/programs/#{program.id}"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Program.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:program) { create(:program, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Program.count).to eq(1)

      delete "/db/programs/#{program.id}"
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Program.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:program) { create(:program, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete program softly" do
      expect(Program.count).to eq(1)

      delete "/db/programs/#{program.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Program.count).to eq(0)
    end
  end
end
