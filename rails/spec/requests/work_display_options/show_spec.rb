# typed: false
# frozen_string_literal: true

RSpec.describe "GET /work_display_option", type: :request do
  it "ログインしているとき、モードを切り替えてリダイレクトすること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect(user.setting.display_option_work_list).to eq "list_detailed"

    get "/work_display_option?display=grid&to=/works/2021-summer"

    expect(user.setting.display_option_work_list).to eq "grid"
    expect(response).to redirect_to("/works/2021-summer?display=grid")
  end

  it "ログインしていないとき、displayパラメータ付きでリダイレクトすること" do
    get "/work_display_option?display=grid&to=/works/2021-summer"

    expect(response).to redirect_to("/works/2021-summer?display=grid")
  end

  it "無効なdisplayパラメータの場合、fallback_locationにリダイレクトすること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/work_display_option?display=invalid&to=root_path", headers: {"HTTP_REFERER" => "/works"}

    expect(user.setting.reload.display_option_work_list).to eq "list_detailed"
    expect(response).to redirect_to("/works")
  end

  it "toパラメータがない場合、displayパラメータ付きでrootにリダイレクトすること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/work_display_option?display=grid"

    expect(user.setting.display_option_work_list).to eq "grid"
    expect(response).to redirect_to("/?display=grid")
  end

  it "ログインユーザーの既存の表示設定と同じdisplayパラメータの場合、設定を更新せずにリダイレクトすること" do
    user = create(:registered_user)
    user.setting.update_column(:display_option_work_list, "grid")
    login_as(user, scope: :user)

    get "/work_display_option?display=grid&to=/works/2021-summer"

    expect(user.setting.display_option_work_list).to eq "grid"
    expect(response).to redirect_to("/works/2021-summer?display=grid")
  end
end
