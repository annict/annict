# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/series", type: :request do
  it "ログインしていない場合、アクセスできずログインページにリダイレクトすること" do
    series_params = {
      name: "シリーズ",
      name_alter: "シリーズ (別名)",
      name_en: "The Series",
      name_alter_en: "The Series (alt)"
    }

    post "/db/series", params: {series: series_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Series.all.size).to eq(0)
  end

  it "エディター権限のないユーザーがログイン時、アクセスできないこと" do
    user = create(:registered_user)
    series_params = {
      name: "シリーズ",
      name_alter: "シリーズ (別名)",
      name_en: "The Series",
      name_alter_en: "The Series (alt)"
    }

    login_as(user, scope: :user)
    post "/db/series", params: {series: series_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Series.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログイン時、シリーズを作成できること" do
    user = create(:registered_user, :with_editor_role)
    series_params = {
      name: "シリーズ",
      name_alter: "シリーズ (別名)",
      name_en: "The Series",
      name_alter_en: "The Series (alt)"
    }

    expect(Series.all.size).to eq(0)

    login_as(user, scope: :user)
    post "/db/series", params: {series: series_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Series.all.size).to eq(1)

    series = Series.first
    expect(series.name).to eq(series_params[:name])
    expect(series.name_alter).to eq(series_params[:name_alter])
    expect(series.name_en).to eq(series_params[:name_en])
    expect(series.name_alter_en).to eq(series_params[:name_alter_en])
  end
end
