# typed: false
# frozen_string_literal: true

describe "GET /work_display_option", type: :request do
  context "ログインしているとき" do
    let!(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    it "モードを切り替えてリダイレクトすること" do
      expect(user.setting.display_option_work_list).to eq "list_detailed"

      get "/work_display_option?display=grid&to=/works/2021-summer"

      expect(user.setting.display_option_work_list).to eq "grid"
      expect(response).to redirect_to("/works/2021-summer?display=grid")
    end
  end
end
