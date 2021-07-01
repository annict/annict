# frozen_string_literal: true

describe "DELETE /db/works/:id", type: :request do
  context "user does not sign in" do
    let!(:work) { create(:anime, :not_deleted) }

    it "user can not access this page" do
      expect(Anime.count).to eq(1)

      delete "/db/works/#{work.id}"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Anime.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:work) { create(:anime, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Anime.count).to eq(1)

      delete "/db/works/#{work.id}"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Anime.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:work) { create(:anime, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Anime.count).to eq(1)

      delete "/db/works/#{work.id}"
      work.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Anime.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:work) { create(:anime, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete work softly" do
      expect(Anime.count).to eq(1)

      delete "/db/works/#{work.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Anime.count).to eq(0)
    end
  end
end
