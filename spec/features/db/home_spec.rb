require "spec_helper"

describe "Annict DB" do
  describe "トップページ" do
    context "ログインしていないとき" do
      before do
        visit "/db"
      end

      it "アクセスするとページが表示される" do
        expect(page).to have_content("Annict DBにようこそ！")
      end
    end

    context "ログインしているとき" do
      let(:user) { create(:registered_user) }

      before do
        login_as(user, scope: :user)
        visit "/db"
      end

      it "アクセスするとページが表示される" do
        expect(page).to have_content("Annict DBにようこそ！")
      end
    end
  end
end
