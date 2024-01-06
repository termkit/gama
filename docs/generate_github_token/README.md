### Generate GitHub Token from GAMA ("Fine-Grained Token")

Generate a GitHub token with specific permissions for optimal use with GAMA. Follow these steps:

1. **Open the Fine-Grained Token Page**
    - Navigate to the fine-grained token page.
    - Click on the "Generate new token" button.

2. **Choose Repositories**
    - Decide the scope of the token: `All repositories` or `Only selected repositories`.
      ![Token Repositories Selection](repos.png)

3. **Set Required Permissions**
    - Scroll through the permissions list and enable the following:

        - **First Permission**: Necessary to trigger workflows.
          ![Permission to Run Workflows](perm1.png)

        - **Second Permission**: Essential to list triggerable workflows.
          ![Permission to Read Workflows](perm2.png)

        - **Third Permission**: Required to read repository contents, enabling workflow triggering.
          ![Permission to Read Repository Contents](perm3.png)

4. **Finalize**
    - After setting the permissions, complete the token generation process.

Now, you can utilize this token with GAMA to manage GitHub Actions workflows effectively.
