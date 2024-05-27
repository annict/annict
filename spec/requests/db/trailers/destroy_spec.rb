# typed: false
# frozen_string_literal: true

describe "DELETE /db/trailers/:id", type: :request do
  context "user does not sign in" do
    let!(:trailer) { create(:trailer, :not_deleted) }

    it "user can not access this page" do
      expect(Trailer.count).to eq(1)

      delete "/db/trailers/#{trailer.id}"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Trailer.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:trailer) { create(:trailer, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Trailer.count).to eq(1)

      delete "/db/trailers/#{trailer.id}"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Trailer.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:trailer) { create(:trailer, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Trailer.count).to eq(1)

      delete "/db/trailers/#{trailer.id}"
      trailer.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Trailer.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:trailer) { create(:trailer, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete trailer softly" do
      expect(Trailer.count).to eq(1)

      delete "/db/trailers/#{trailer.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Trailer.count).to eq(0)
    end
  end
end
