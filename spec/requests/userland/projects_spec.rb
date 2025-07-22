# typed: false
# frozen_string_literal: true

RSpec.describe "POST /userland/projects", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    category = FactoryBot.create(:userland_category)

    post "/userland/projects", params: {
      userland_project: {
        userland_category_id: category.id,
        name: "テストプロジェクト",
        summary: "テストプロジェクトの概要",
        description: "テストプロジェクトの詳細説明",
        url: "https://example.com",
        available: true
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、有効なパラメータでプロジェクトが作成されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "テストプロジェクトの概要",
          description: "テストプロジェクトの詳細説明",
          url: "https://example.com",
          available: true
        }
      }
    }.to change(UserlandProject, :count).by(1)

    project = UserlandProject.last
    expect(project.name).to eq("テストプロジェクト")
    expect(project.summary).to eq("テストプロジェクトの概要")
    expect(project.description).to eq("テストプロジェクトの詳細説明")
    expect(project.url).to eq("https://example.com")
    expect(project.available).to be true
    expect(project.userland_category).to eq(category)
    expect(project.users).to include(user)

    expect(response).to redirect_to(userland_project_path(project))
    expect(flash[:notice]).to eq(I18n.t("messages._common.created"))
  end

  it "ログインしているとき、名前が空の場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "",
          summary: "テストプロジェクトの概要",
          description: "テストプロジェクトの詳細説明",
          url: "https://example.com",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、概要が空の場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "",
          description: "テストプロジェクトの詳細説明",
          url: "https://example.com",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、詳細説明が空の場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "テストプロジェクトの概要",
          description: "",
          url: "https://example.com",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、URLが空の場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "テストプロジェクトの概要",
          description: "テストプロジェクトの詳細説明",
          url: "",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、URLが無効な形式の場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "テストプロジェクトの概要",
          description: "テストプロジェクトの詳細説明",
          url: "invalid-url",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、名前が長すぎる場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "a" * 51,
          summary: "テストプロジェクトの概要",
          description: "テストプロジェクトの詳細説明",
          url: "https://example.com",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、概要が長すぎる場合バリデーションエラーでnewテンプレートが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)

    login_as(user, scope: :user)

    expect {
      post "/userland/projects", params: {
        userland_project: {
          userland_category_id: category.id,
          name: "テストプロジェクト",
          summary: "a" * 151,
          description: "テストプロジェクトの詳細説明",
          url: "https://example.com",
          available: true
        }
      }
    }.not_to change(UserlandProject, :count)

    expect(response.status).to eq(200)
  end
end
