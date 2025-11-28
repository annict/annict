# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /userland/projects/:project_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    category = create(:userland_category)
    project = create(:userland_project, userland_category: category)

    patch "/userland/projects/#{project.id}", params: {
      userland_project: {
        name: "更新されたプロジェクト名",
        summary: "更新された概要"
      }
    }

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "プロジェクトメンバーでないユーザーの場合、アクセスできないこと" do
    user = create(:registered_user)
    category = create(:userland_category)
    project = create(:userland_project, userland_category: category)
    login_as(user, scope: :user)

    expect {
      patch "/userland/projects/#{project.id}", params: {
        userland_project: {
          name: "更新されたプロジェクト名",
          summary: "更新された概要"
        }
      }
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "プロジェクトメンバーの場合、プロジェクトを更新できること" do
    user = create(:registered_user)
    category = create(:userland_category)
    project = create(:userland_project, userland_category: category)
    create(:userland_project_member, user: user, userland_project: project)
    login_as(user, scope: :user)

    patch "/userland/projects/#{project.id}", params: {
      userland_project: {
        name: "更新されたプロジェクト名",
        summary: "更新された概要"
      }
    }

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    project.reload
    expect(project.name).to eq("更新されたプロジェクト名")
    expect(project.summary).to eq("更新された概要")
  end

  it "バリデーションエラーがある場合、editテンプレートを再表示すること" do
    user = create(:registered_user)
    category = create(:userland_category)
    project = create(:userland_project, userland_category: category)
    create(:userland_project_member, user: user, userland_project: project)
    login_as(user, scope: :user)

    patch "/userland/projects/#{project.id}", params: {
      userland_project: {
        name: "",
        summary: ""
      }
    }

    expect(response.status).to eq(200)
    expect(response.body).to include("編集")
  end

  it "存在しないプロジェクトIDの場合、404エラーになること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      patch "/userland/projects/999999", params: {
        userland_project: {
          name: "更新されたプロジェクト名",
          summary: "更新された概要"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
