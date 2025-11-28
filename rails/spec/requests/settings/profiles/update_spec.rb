# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /settings/profile", type: :request do
  it "ログイン済みユーザーがプロフィール名を更新できること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        name: "新しい名前"
      }
    }

    expect(response).to redirect_to(settings_profile_path)
    expect(user.profile.reload.name).to eq("新しい名前")
  end

  it "ログイン済みユーザーがプロフィール説明を更新できること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        description: "新しい説明文です"
      }
    }

    expect(response).to redirect_to(settings_profile_path)
    expect(user.profile.reload.description).to eq("新しい説明文です")
  end

  it "ログイン済みユーザーがプロフィールURLを更新できること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        url: "https://example.com"
      }
    }

    expect(response).to redirect_to(settings_profile_path)
    expect(user.profile.reload.url).to eq("https://example.com")
  end

  it "ログイン済みユーザーが複数の属性を同時に更新できること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        name: "更新した名前",
        description: "更新した説明文",
        url: "https://updated.example.com"
      }
    }

    expect(response).to redirect_to(settings_profile_path)
    profile = user.profile.reload
    expect(profile.name).to eq("更新した名前")
    expect(profile.description).to eq("更新した説明文")
    expect(profile.url).to eq("https://updated.example.com")
  end

  it "プロフィール名が空の場合、更新が失敗すること" do
    user = FactoryBot.create(:registered_user)
    original_name = user.profile.name
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        name: ""
      }
    }

    expect(response.status).to eq(422)
    expect(user.profile.reload.name).to eq(original_name)
  end

  it "プロフィール説明が150文字を超える場合、150文字に切り詰められること" do
    user = FactoryBot.create(:registered_user)
    login_as(user, scope: :user)
    long_description = "a" * 200

    patch "/settings/profile", params: {
      profile: {
        description: long_description
      }
    }

    expect(response).to redirect_to(settings_profile_path)
    expect(user.profile.reload.description.length).to eq(150)
    expect(user.profile.reload.description).to eq("a" * 147 + "...")
  end

  it "無効なURL形式の場合、更新が失敗すること" do
    user = FactoryBot.create(:registered_user)
    original_url = user.profile.url
    login_as(user, scope: :user)

    patch "/settings/profile", params: {
      profile: {
        url: "invalid-url"
      }
    }

    expect(response.status).to eq(422)
    expect(user.profile.reload.url).to eq(original_url)
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    patch "/settings/profile", params: {
      profile: {
        name: "新しい名前"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end
end
