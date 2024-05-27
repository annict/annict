# typed: false
# frozen_string_literal: true

describe "DELETE /db/series/:id", type: :request do
  context "user does not sign in" do
    let!(:series) { create(:series, :not_deleted) }

    it "user can not access this page" do
      expect(Series.count).to eq(1)

      delete "/db/series/#{series.id}"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Series.count).to eq(1)
    end
  end

  context "user who is not editor signs in" do
    let!(:user) { create(:registered_user) }
    let!(:series) { create(:series, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Series.count).to eq(1)

      delete "/db/series/#{series.id}"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Series.count).to eq(1)
    end
  end

  context "user who is editor signs in" do
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:series) { create(:series, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      expect(Series.count).to eq(1)

      delete "/db/series/#{series.id}"
      series.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Series.count).to eq(1)
    end
  end

  context "user who is admin signs in" do
    let!(:user) { create(:registered_user, :with_admin_role) }
    let!(:series) { create(:series, :not_deleted) }

    before do
      login_as(user, scope: :user)
    end

    it "user can delete series softly" do
      expect(Series.count).to eq(1)

      delete "/db/series/#{series.id}"

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("削除しました")

      expect(Series.count).to eq(0)
    end
  end
end
