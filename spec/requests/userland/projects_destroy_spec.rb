# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /userland/projects/:project_id", type: :request do
  it "認証されていないユーザーがアクセスすると、サインインページにリダイレクトされること" do
    userland_project = create(:userland_project)

    delete "/userland/projects/#{userland_project.id}"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "認証されているが権限のないユーザーがアクセスすると、Pundit::NotAuthorizedErrorが発生すること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    userland_project = create(:userland_project)
    create(:userland_project_member, userland_project: userland_project, user: other_user)

    login_as(user, scope: :user)

    expect do
      delete "/userland/projects/#{userland_project.id}"
    end.to raise_error(Pundit::NotAuthorizedError)
  end

  it "権限のあるユーザーがプロジェクトを削除できること" do
    user = create(:registered_user)
    userland_project = create(:userland_project)
    create(:userland_project_member, userland_project: userland_project, user: user)

    login_as(user, scope: :user)

    expect do
      delete "/userland/projects/#{userland_project.id}"
    end.to change(UserlandProject, :count).by(-1)

    expect(response).to redirect_to(userland_path)
    expect(flash[:notice]).to eq(I18n.t("messages._common.deleted"))
  end

  it "存在しないプロジェクトIDでアクセスするとActionController::RoutingErrorが発生すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    delete "/userland/projects/non_existent_id"

    expect(response.status).to eq(404)
  end

  it "有効なID形式だが存在しないプロジェクトIDでアクセスすると404エラーが返されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    delete "/userland/projects/999999"

    expect(response).to have_http_status(:not_found)
  end

  it "プロジェクトが削除されると関連するプロジェクトメンバーも削除されること" do
    user = create(:registered_user)
    userland_project = create(:userland_project)
    project_member = create(:userland_project_member, userland_project: userland_project, user: user)

    login_as(user, scope: :user)

    expect do
      delete "/userland/projects/#{userland_project.id}"
    end.to change(UserlandProjectMember, :count).by(-1)

    expect(UserlandProjectMember.find_by(id: project_member.id)).to be_nil
  end

  it "プロジェクト削除時に適切なflashメッセージが設定されること" do
    user = create(:registered_user)
    userland_project = create(:userland_project)
    create(:userland_project_member, userland_project: userland_project, user: user)

    login_as(user, scope: :user)
    delete "/userland/projects/#{userland_project.id}"

    expect(flash[:notice]).to eq(I18n.t("messages._common.deleted"))
  end
end
