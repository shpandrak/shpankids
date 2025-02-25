package webserver

const IndexPage = `
<html>
	<head>
		<title>Shpankids Login</title>
		<style>
			body {
				background-color: #0f2b3e;
				color: #ffffff;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				font-family: Arial, sans-serif;
			}
			.container {
				background: rgba(255, 255, 255, 0.1);
				padding: 20px;
				border-radius: 8px;
				text-align: center;
				box-shadow: 0 0 10px rgba(0, 0, 0, 0.2);
			}
			h2 {
				margin-bottom: 10px;
			}
			ul {
				list-style: none;
				padding: 0;
			}
			li {
				margin: 10px 0;
			}
			a {
				color: #646cff;
				text-decoration: none;
				font-weight: bold;
			}
			a:hover {
				text-decoration: underline;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h2>Shpankids Login</h2>
			<p>"Select" a login method:</p>
			<ul>
				<li><a href="/login-gl">Login with Google</a></li>
			</ul>
		</div>
	</body>
</html>
`
